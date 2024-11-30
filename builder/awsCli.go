package builder

// FOR this to work you need to have aws cli installed
// esto funciono en el cli
// aws s3 cp test.txt s3://fileserver-test-5fe07c2485974f7385d6624-eba3330.divio-media.com/test.txt --acl public-read
// fgs@MacBook-Pro-de-Francisco Downloads % aws s3 ls s3://fileserver-test-5fe07c2485974f7385d6624-eba3330.divio-media.com --recursive | awk '{print $4}'
// uploads
// xxx.env

// aws s3 cp cors.json s3://fileserver-test-5fe07c2485974f7385d6624-eba3330.divio-media.com/cors.json
// aws s3api put-bucket-cors --bucket fileserver-test-5fe07c2485974f7385d6624-eba3330.divio-media.com --cors-configuration file://cors.json

// aws s3api get-bucket-cors --bucket fileserver-test-5fe07c2485974f7385d6624-eba3330.divio-media.com

// aws s3api delete-object --bucket fileserver-test-5fe07c2485974f7385d6624-eba3330.divio-media.com --key test.txt

// main uses the AWS SDK for Go V2 to create an Amazon Simple Sto

// func awsReady() {
// 	var stdout bytes.Buffer
// 	var stderr bytes.Buffer
// 	cmd := exec.Command(ShellToUse, "-c", command)
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr
// 	err := cmd.Run()

// }

// func execCommand(command string) {
// 	// aws s3 cp test.txt s3://fileserver-test-5fe07c2485974f7385d6624-eba3330.divio-media.com/test.txt --acl public-read
// }
