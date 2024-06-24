package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
	"rssagg/internal/database"
	"time"
)

func (apiCfg *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	type params struct {
		FeedID uuid.UUID `json:"feed_id"`
	}
	p := params{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	feedFollow, err := apiCfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    p.FeedID,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Could not create feed follow: %s", err.Error()))
		return
	}
	respondWithJSON(w, http.StatusCreated, feedFollow)
}

func (apiCfg *apiConfig) handlerGetFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	FeedFollowIDParam := chi.URLParam(r, "FeedFollowID")
	FeedFollowID, err := uuid.Parse(FeedFollowIDParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid feed follow ID")
		return
	}
	feedFollow, err := apiCfg.DB.GetFeedFollow(r.Context(), FeedFollowID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Feed follow not found")
		return
	}
	respondWithJSON(w, http.StatusOK, feedFollow)
}

func (apiCfg *apiConfig) handlerGetFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollows, err := apiCfg.DB.GetFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, feedFollows)
}

func (apiCfg *apiConfig) handlerDeleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	FeedFollowIDParam := chi.URLParam(r, "FeedFollowID")
	FeedFollowID, err := uuid.Parse(FeedFollowIDParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid feed follow ID")
		return
	}
	feedFollow, err := apiCfg.DB.GetFeedFollow(r.Context(), FeedFollowID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if feedFollow.UserID != user.ID {
		respondWithError(w, http.StatusForbidden, "You cannot delete this follow")
		return
	}
	err = apiCfg.DB.DeleteFeedFollow(r.Context(), FeedFollowID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusNoContent, "")
}
