package main

import "github.com/daemtri/begonia/logx"

func main() {
	logger := logx.GetLogger("test")
	logger.Info("hello world")
}
