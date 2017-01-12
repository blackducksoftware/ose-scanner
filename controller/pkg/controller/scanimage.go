package controller

import (
	"log"
	"strings"
)

type ScanImage struct {
	imageId		string
	taggedName 	string
	digest		string
	scanned		bool
}

func NewScanImage (ID string, Reference string) *ScanImage  {

	tag := strings.Split(Reference, "@")

	return &ScanImage {
		imageId: ID,
		taggedName: tag[0],
		digest: Reference,
		scanned: false,
	}
}

func (image ScanImage) scan () (e error){

	log.Printf ("Scanning: %s (%s)\n", image.taggedName, image.imageId[:10])

	args := []string {}
	args = append(args, "/ose_scanner")

	args = append(args, "-h")
	args = append(args, Hub.Host)

	args = append(args, "-p")
	args = append(args, Hub.Port)	

	args = append(args, "-s")
	args = append(args, Hub.Scheme)

	args = append(args, "-u")
	args = append(args, Hub.Username)

	args = append(args, "-w")
	args = append(args, Hub.Password)

	args = append(args, "-id")
	args = append(args, image.imageId)

	args = append(args, "-tag")
	args = append(args, image.taggedName)

	args = append(args, "-digest")
	args = append(args, image.digest)

	docker := NewDocker ()

	err := docker.launchContainer (Hub.Scanner, args)

	if err != nil {
		log.Printf ("Error creating scanning container: %s\n", err)
		return err
	}
	

	log.Printf ("Done Scanning: %s\n", image.taggedName)

	image.scanned = true

	return nil
}

