package jwtc

import (
	"context"
	"os"
	"testing"

	sctx "inventory-movement-processing/pkg/service_context"
)

var testServiceCtx sctx.ServiceContext

func TestMain(m *testing.M) {
	_ = os.Setenv("JWT_SECRET", "this-is-a-very-secure-secret-key-32bytes!")
	_ = os.Setenv("JWT_EXP_SECS", "60")

	testServiceCtx = sctx.NewServiceContext(
		sctx.WithName("test"),
		sctx.WithComponent(NewJWT("jwt")),
	)

	if err := testServiceCtx.Load(); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = testServiceCtx.Stop()
	os.Exit(code)
}

func TestJWT_IssueToken_Success(t *testing.T) {
	j := testServiceCtx.MustGet("jwt").(*jwtx)

	token, exp, err := j.IssueToken(
		context.Background(),
		"token-id-1",
		"user-1",
		0,
	)

	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}

	if exp != j.expireTokenInSeconds {
		t.Fatalf("expected exp=%d, got %d", j.expireTokenInSeconds, exp)
	}
}

func TestJWT_ParseToken_Success(t *testing.T) {
	j := testServiceCtx.MustGet("jwt").(*jwtx)

	token, _, err := j.IssueToken(
		context.Background(),
		"token-id-1",
		"user-1",
		0,
	)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := j.ParseToken(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}

	if claims.Subject != "user-1" {
		t.Fatalf("unexpected subject: %s", claims.Subject)
	}

	if claims.ID != "token-id-1" {
		t.Fatalf("unexpected token id: %s", claims.ID)
	}
}

func TestJWT_ParseToken_InvalidToken(t *testing.T) {
	j := testServiceCtx.MustGet("jwt").(*jwtx)

	token, _, err := j.IssueToken(
		context.Background(),
		"id",
		"user",
		0,
	)
	if err != nil {
		t.Fatal(err)
	}

	badToken := token + "broken"

	_, err = j.ParseToken(context.Background(), badToken)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
