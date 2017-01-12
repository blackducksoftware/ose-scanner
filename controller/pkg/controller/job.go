package controller

import (
	"log"
)

type Job struct {
	ScanImage 	*ScanImage
	controller	*Controller
}


func (job Job) Done() {
	job.controller.wait.Done()
	return
}

func (job Job) Load() {
	job.controller.wait.Add(1)
	log.Println ("Queue image: " + job.ScanImage.taggedName)
	return
}

