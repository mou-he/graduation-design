package image

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
	"golang.org/x/image/draw"
)

type ImageRecognizer struct {
	session      *ort.Session[float32] //onnx推理会话
	inputName    string                //输入张量名称
	outputName   string                //输出张量名称
	inputHeight  int                   //输入张量高度
	inputWidth   int                   //输入张量宽度
	labels       []string              //分类标签
	inputTensor  *ort.Tensor[float32]  //输入张量
	outputTensor *ort.Tensor[float32]  //输出张量
}

const (
	defaultInputName  = "data"
	defaultOutputName = "mobilenetv20_output_flatten0_reshape0"
)

var (
	initOnce sync.Once
	initErr  error
)

// 创建识别器
func NewImageRecognizer(modelPath string, labelPath string, inputHeight, inputWeight int) (*ImageRecognizer, error) {
	// 初始化默认尺寸
	if inputHeight <= 0 || inputWeight <= 0 {
		inputHeight, inputWeight = 224, 224
	}
	// 初始化ONNX环境
	initOnce.Do(func() {
		initErr = ort.InitializeEnvironment()
	})
	if initErr != nil {
		return nil, fmt.Errorf("init onnx environment failed: %v", initErr)
	}
	// 预先创建输入输出的Tensor
	inputShape := ort.NewShape(int64(1), int64(3), int64(inputHeight), int64(inputWeight))
	inData := make([]float32, inputShape.FlattenedSize())
	inTensor, err := ort.NewTensor(inputShape, inData)
	if err != nil {
		return nil, fmt.Errorf("create input tensor failed: %v", err)
	}
	// 创建输出张量
	outShape := ort.NewShape(1, 1000)
	outTensor, err := ort.NewEmptyTensor[float32](outShape)
	if err != nil {
		return nil, fmt.Errorf("create output tensor failed: %v", err)
	}
	// 创建会话
	session, err := ort.NewSession[float32](
		modelPath,
		[]string{defaultInputName},
		[]string{defaultOutputName},
		[]*ort.Tensor[float32]{inTensor},
		[]*ort.Tensor[float32]{outTensor},
	)
	if err != nil {
		inTensor.Destroy()
		outTensor.Destroy()
		return nil, fmt.Errorf("create session failed: %v", err)
	}
	// 读取label文件
	labels, err := loadLabels(labelPath)
	if err != nil {
		session.Destroy()
		inTensor.Destroy()
		outTensor.Destroy()
		return nil, fmt.Errorf("load labels failed: %v", err)
	}
	return &ImageRecognizer{
		session:      session,
		inputName:    defaultInputName,
		outputName:   defaultOutputName,
		inputHeight:  inputHeight,
		inputWidth:   inputWeight,
		labels:       labels,
		inputTensor:  inTensor,
		outputTensor: outTensor,
	}, nil

}
func (r ImageRecognizer) Close() {
	if r.session != nil {
		_ = r.session.Destroy()
		r.session = nil
	}
	if r.inputTensor != nil {
		_ = r.inputTensor.Destroy()
		r.inputTensor = nil
	}
	if r.outputTensor != nil {
		_ = r.outputTensor.Destroy()
		r.outputTensor = nil
	}
}
func (r *ImageRecognizer) PredictFromFile(imagePath string) (string, error) {
	file, err := os.Open(filepath.Clean(imagePath))
	if err != nil {
		return "", fmt.Errorf("open image file failed: %v", err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("decode image failed: %v", err)
	}
	return r.PredictFromImage(img)
}
func (r *ImageRecognizer) PredictFromBuffer(buf []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("decode image failed: %v", err)
	}
	return r.PredictFromImage(img)
}
func (r *ImageRecognizer) PredictFromImage(img image.Image) (string, error) {
	// 调整图像大小
	resizedImg := image.NewRGBA(image.Rect(0, 0, r.inputWidth, r.inputHeight))
	// 调整图像大小
	draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	// 转化为模型输入
	h, w := r.inputHeight, r.inputWidth
	ch := 3 // R, G, B
	data := make([]float32, h*w*ch)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// 获取像素值
			c := resizedImg.At(x, y)
			// 归一化
			r, g, b, _ := c.RGBA()
			// 归一化到[0, 1]
			rf := float32(r>>8) / 255.0
			// 归一化到[0, 1]
			gf := float32(g>>8) / 255.0
			// 归一化到[0, 1]
			bf := float32(b>>8) / 255.0

			//R通道
			data[y*w+x] = rf
			// G通道
			data[h*w+y*w+x] = gf
			// B通道
			data[2*h*w+y*w+x] = bf
		}
	}
	inData := r.inputTensor.GetData()
	// 复制数据到输入张量
	copy(inData, data)
	if err := r.session.Run(); err != nil {
		return "", fmt.Errorf("run session failed: %v", err)
	}
	outData := r.outputTensor.GetData()
	if len(outData) == 0 {
		return "", fmt.Errorf("output tensor is empty")
	}
	maxIdx := 0
	maxVal := outData[0]
	for i := 1; i < len(outData); i++ {
		if outData[i] > maxVal {
			maxVal = outData[i]
			maxIdx = i
		}
	}
	if maxIdx >= 0 && maxIdx < len(r.labels) {
		return r.labels[maxIdx], nil
	}
	return "unknown", nil
}
func loadLabels(labelPath string) ([]string, error) {
	// 读取Label文件
	f, err := os.Open(filepath.Clean(labelPath))
	if err != nil {
		return nil, fmt.Errorf("open label file failed: %v", err)
	}
	defer f.Close()
	var labels []string
	//
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			labels = append(labels, line)
		}
	}
	// 检查扫描错误
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan label file failed: %v", err)
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("label file is empty")
	}
	return labels, nil
}
