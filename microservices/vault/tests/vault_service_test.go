package tests

import (
	"context"
	"testing"

	"github.com/ValeryPiashchynski/TaskManager/microservices/tools"
	"github.com/ValeryPiashchynski/TaskManager/microservices/vault"
)

func TestVaultService(t *testing.T) {
	pswd := tools.NewPasswordChecker()
	srv := vault.newVaultService(pswd)
	ctx := context.Background()
	h, err := srv.Hash(ctx, "password")
	if err != nil {
		t.Error(err)
	}
	ok, err := srv.Validate(ctx, "password", h)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Hashes mismatch")
	}

	okk, err := srv.Validate(ctx, "wrong_password", h)
	if err == nil {
		t.Errorf("Valid %s ", err)
	}
	if okk {
		t.Error("Expected false from valid")
	}
}
