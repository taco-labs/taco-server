package driver

import (
	"context"
	"fmt"

	"github.com/taco-labs/taco/go/domain/value"
	"github.com/taco-labs/taco/go/domain/value/enum"
	"golang.org/x/sync/errgroup"
)

func (d driverApp) driverImageUrls(ctx context.Context, driverId string) (value.DriverImageUrls, value.DriverImageUrls, error) {
	var driverProfileDownloadUrl, driverProfileUploadUrl, driverLicenseDownloadUrl, driverLicenseUploadUrl string
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		url, err := d.service.imageDownloadUrl.GetDownloadUrl(ctx, getImageFileName(driverId, enum.ImageType_DriverProfile))
		if err != nil {
			return fmt.Errorf("app.driver.GetDriverImageUrls: error while get driver profile download url:%w", err)
		}
		driverProfileDownloadUrl = url
		return nil
	})
	group.Go(func() error {
		url, err := d.service.imageUploadUrl.GetUploadUrl(ctx, getImageFileName(driverId, enum.ImageType_DriverProfile))
		if err != nil {
			return fmt.Errorf("app.driver.GetDriverImageUrls: error while get driver profile upload url:%w", err)
		}
		driverProfileUploadUrl = url
		return nil
	})
	group.Go(func() error {
		url, err := d.service.imageDownloadUrl.GetDownloadUrl(ctx, getImageFileName(driverId, enum.ImageType_DriverLicense))
		if err != nil {
			return fmt.Errorf("app.driver.GetDriverImageUrls: error while get licnese image download url:%w", err)
		}
		driverLicenseDownloadUrl = url
		return nil
	})
	group.Go(func() error {
		url, err := d.service.imageUploadUrl.GetUploadUrl(ctx, getImageFileName(driverId, enum.ImageType_DriverLicense))
		if err != nil {
			return fmt.Errorf("app.driver.GetDriverImageUrls: error while get license image upload url:%w", err)
		}
		driverLicenseUploadUrl = url
		return nil
	})
	if err := group.Wait(); err != nil {
		return value.DriverImageUrls{}, value.DriverImageUrls{}, err
	}

	downloadUrls := value.DriverImageUrls{
		ProfileImage: driverProfileDownloadUrl,
		LicenseImage: driverLicenseDownloadUrl,
	}
	uploadUrls := value.DriverImageUrls{
		ProfileImage: driverProfileUploadUrl,
		LicenseImage: driverLicenseUploadUrl,
	}

	return downloadUrls, uploadUrls, nil
}
