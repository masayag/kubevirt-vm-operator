package utils

var osVersionMap = map[string]string{
	"29": "kubevirt/fedora-cloud-registry-disk-demo",
	"30": "kubevirt/fedora-cloud-registry-disk-demo",
}

// GetFedoraImage returns the image for Fedora based on the required version
// TODO: add webhook validation for the existance if an image for the specific version
// TODO: implement mapping between the OS version to the image
func GetFedoraImage(OSVersion string) string {
	return osVersionMap[OSVersion]
}
