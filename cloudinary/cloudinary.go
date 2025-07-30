package cloudinary

import "github.com/cloudinary/cloudinary-go/v2"

func NewCloudinaryInstance() (*cloudinary.Cloudinary, error) {

	cld, err := cloudinary.New()
	if err != nil {
		return nil, err
	}

	return cld, nil
}
