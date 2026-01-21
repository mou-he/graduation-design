package image

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	ort "github.com/yalue/onnxruntime_go"
)

type imageRecognizer struct {
	session      *ort.Session[float32]
	inputName    string
	outputName   string
	inputHeight  int
	inputWidth   int
	labels       []string
	inputTensor  *ort.Tensor[float32]
	outputTensor *ort.Tensor[float32]
}
