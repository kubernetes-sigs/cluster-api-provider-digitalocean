package computes

import (
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha3"
)

// GetVolumeByName takes a volume name and returns a Volume if found.
func (s *Service) GetVolumeByName(name string) (*godo.Volume, error) {
	vols, _, err := s.scope.Storage.ListVolumes(s.ctx, &godo.ListVolumeParams{
		Name:   name,
		Region: s.scope.Region(),
	})
	if err != nil {
		return nil, TemporaryError{fmt.Errorf("failed to list volumes: %w", err)}
	}
	if len(vols) == 0 {
		return nil, nil
	}
	if len(vols) > 1 {
		return nil, errors.New("volume names are not unique per region")
	}
	return &vols[0], nil
}

// CreateVolume creates a block storage volume.
func (s *Service) CreateVolume(disk infrav1.DataDisk, volName string) (*godo.Volume, error) {
	r := &godo.VolumeCreateRequest{
		Region:          s.scope.Region(),
		Name:            volName,
		SizeGigaBytes:   disk.DiskSizeGB,
		FilesystemType:  disk.FilesystemType,
		FilesystemLabel: disk.FilesystemLabel,
	}
	v, _, err := s.scope.Storage.CreateVolume(s.ctx, r)
	return v, errors.Wrap(err, "failed to create new volume")
}

// DeleteVolume deletes a block storage volume.
func (s *Service) DeleteVolume(id string) error {
	s.scope.V(2).Info("Attempting to delete block storage volume", "volume-id", id)

	if resp, err := s.scope.Storage.DeleteVolume(s.ctx, id); err != nil {
		if resp.StatusCode >= 500 {
			err = TemporaryError{err}
		}
		return fmt.Errorf("failed to delete instance with id %q: %w", id, err)
	}

	s.scope.V(2).Info("Deleted block storage volume", "volume-id", id)
	return nil
}
