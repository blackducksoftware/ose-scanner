package controller

import (
	"fmt"
)

type Job struct {
	ScanImage 	ScanImage
	controller	*Controller
}


func (job Job) Done() {
	job.controller.wait.Done()
	return
}

func (job Job) Load() {
	job.controller.wait.Add(1)
	fmt.Println ("Queue image: " + job.ScanImage.taggedName)
	return
}

