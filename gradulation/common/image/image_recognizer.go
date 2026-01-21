package image

import (
	"bufio"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
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
		initErr = ort.InitGlobalSessionEnvironment()
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
