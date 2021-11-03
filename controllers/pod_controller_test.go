package controllers

import (
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestSplitImage(t *testing.T) {

	image := "harbor.mytanzu.xyz/library/micropet-tap-pets:latest"

	domain, repository, tag := splitImage(image)

	assertEqual(t, "harbor.mytanzu.xyz", domain)
	assertEqual(t, "library/micropet-tap-pets", repository)
	assertEqual(t, "latest", tag)
}

func TestSplitImageSha256(t *testing.T) {

	image := "harbor.mytanzu.xyz/library/micropet-tap-pets@sha256:446be1d21a57a6e92312e10a7530bd5da34240e80f0855a03061d2dabd479177"

	domain, repository, tag := splitImage(image)

	assertEqual(t, "harbor.mytanzu.xyz", domain)
	assertEqual(t, "library/micropet-tap-pets", repository)
	assertEqual(t, "sha256:446be1d21a57a6e92312e10a7530bd5da34240e80f0855a03061d2dabd479177", tag)
}
