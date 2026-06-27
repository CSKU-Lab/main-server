package converter

import (
	"fmt"

	"github.com/CSKU-Lab/main-server/configs"
)

func ToS3Path(config *configs.Config, path string) string {
	return fmt.Sprintf("%s/%s", config.S3_Frontend_URL, path)
}
