package user

type UpdateMeRequest struct {
	Email    *string `json:"email" binding:"omitempty,email,max=255"`
	Password *string `json:"password" binding:"omitempty,min=8,max=72"`
}

type AdminUpdateRequest struct {
	Email    *string `json:"email" binding:"omitempty,email,max=255"`
	Password *string `json:"password" binding:"omitempty,min=8,max=72"`
	Role     *string `json:"role" binding:"omitempty,oneof=user admin"`
}
