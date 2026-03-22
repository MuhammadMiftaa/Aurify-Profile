package service

import (
	"context"
	"fmt"
	"time"

	"refina-profile/config/miniofs"
	"refina-profile/internal/repository"
	"refina-profile/internal/types/dto"
	"refina-profile/internal/types/model"

	"github.com/google/uuid"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID string) (*dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	UploadPhoto(ctx context.Context, userID string, base64Image string) (*dto.UploadPhotoResponse, error)
	DeletePhoto(ctx context.Context, userID string) (*dto.DeletePhotoResponse, error)
}

type profileService struct {
	repo  repository.ProfileRepository
	minio *miniofs.MinIOManager
}

func NewProfileService(repo repository.ProfileRepository, minio *miniofs.MinIOManager) ProfileService {
	return &profileService{
		repo:  repo,
		minio: minio,
	}
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*dto.ProfileResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	profile, err := s.repo.GetByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// If profile doesn't exist, create a new one
	if profile == nil {
		profile = &model.Profile{
			UserID:   userUUID,
			Fullname: "",
			PhotoURL: "",
		}
		if err := s.repo.Create(ctx, profile); err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
	}

	return toProfileResponse(profile), nil
}

func (s *profileService) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	profile, err := s.repo.GetByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// If profile doesn't exist, create a new one
	if profile == nil {
		profile = &model.Profile{
			UserID:   userUUID,
			Fullname: req.Fullname,
			PhotoURL: "",
		}
		if err := s.repo.Create(ctx, profile); err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
	} else {
		// Update existing profile
		profile.Fullname = req.Fullname
		if err := s.repo.Update(ctx, profile); err != nil {
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}
	}

	return toProfileResponse(profile), nil
}

func (s *profileService) UploadPhoto(ctx context.Context, userID string, base64Image string) (*dto.UploadPhotoResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	profile, err := s.repo.GetByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// If profile doesn't exist, create a new one
	if profile == nil {
		profile = &model.Profile{
			UserID:   userUUID,
			Fullname: "",
			PhotoURL: "",
		}
		if err := s.repo.Create(ctx, profile); err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
	}

	// Delete old photo if exists
	if profile.PhotoURL != "" {
		objectName, err := s.minio.ExtractObjectName(profile.PhotoURL)
		if err == nil {
			_ = s.minio.DeleteObject(ctx, "", objectName)
		}
	}

	// Upload new photo
	uploadResp, err := s.minio.UploadBase64(ctx, miniofs.UploadRequest{
		Base64Data: base64Image,
		Prefix:     fmt.Sprintf("profile_%s", userID),
		Validation: &miniofs.FileValidationConfig{
			AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".webp"},
			MaxFileSize:       5 * 1024 * 1024, // 5MB
			MinFileSize:       1024,            // 1KB
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload photo: %w", err)
	}

	// Update profile with new photo URL
	profile.PhotoURL = uploadResp.URL
	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &dto.UploadPhotoResponse{
		Success:  true,
		PhotoURL: uploadResp.URL,
		Message:  "Profile photo uploaded successfully",
	}, nil
}

func (s *profileService) DeletePhoto(ctx context.Context, userID string) (*dto.DeletePhotoResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	profile, err := s.repo.GetByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if profile == nil {
		return &dto.DeletePhotoResponse{
			Success: true,
			Message: "No profile photo to delete",
		}, nil
	}

	// Delete photo from MinIO
	if profile.PhotoURL != "" {
		objectName, err := s.minio.ExtractObjectName(profile.PhotoURL)
		if err == nil {
			if err := s.minio.DeleteObject(ctx, "", objectName); err != nil {
				return nil, fmt.Errorf("failed to delete photo: %w", err)
			}
		}
	}

	// Update profile to remove photo URL
	profile.PhotoURL = ""
	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &dto.DeletePhotoResponse{
		Success: true,
		Message: "Profile photo deleted successfully",
	}, nil
}

func toProfileResponse(profile *model.Profile) *dto.ProfileResponse {
	return &dto.ProfileResponse{
		ID:        profile.ID.String(),
		UserID:    profile.UserID.String(),
		Fullname:  profile.Fullname,
		PhotoURL:  profile.PhotoURL,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt: profile.UpdatedAt.Format(time.RFC3339),
	}
}
