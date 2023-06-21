package main

import "git.bianfeng.com/stars/wegame/wan/wanx/logx"

func main() {
	logger := logx.GetLogger("test")
	logger.Info("hello world")
}
