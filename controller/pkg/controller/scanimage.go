package controller

import (
	"fmt"
)

type ScanImage struct {
	imageId		string
	taggedName 	string
	digest		string
}


func (image ScanImage) scan () (e error){

	//cmd := fmt.Sprintf ("docker run -ti --rm -v /var/run/docker.sock:/var/run/docker.sock --privileged %s /ose_scanner -h %s -p %s -s %s -u %s -w %s -id %s -tag %s", Hub.Scanner, Hub.Host, Hub.Port, Hub.Scheme, Hub.Username, Hub.Password, image.imageId, image.taggedName)

	fmt.Printf ("Scanning: %s (%s)\n", image.taggedName, image.imageId[:10])

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
		fmt.Printf ("Error creating scanning container: %s\n", err)
		return err
	}

	fmt.Printf ("Done Scanning: %s\n", image.taggedName)

	//time.Sleep(10*time.Second)
	return nil
}

