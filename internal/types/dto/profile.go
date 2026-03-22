package dto

type ProfileResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Fullname  string `json:"fullname"`
	PhotoURL  string `json:"photo_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UpdateProfileRequest struct {
	Fullname string `json:"fullname"`
}

type UploadPhotoRequest struct {
	Base64Image string `json:"base64_image" validate:"required"`
}

type UploadPhotoResponse struct {
	Success  bool   `json:"success"`
	PhotoURL string `json:"photo_url"`
	Message  string `json:"message"`
}

type DeletePhotoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
