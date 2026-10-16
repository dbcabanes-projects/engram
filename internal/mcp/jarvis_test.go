package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// Copia de JARVIS: el portátil entra por SSH con `engram mcp --project <repo>`, y un repo nuevo
// todavía no tiene memorias. Leer ese proyecto tiene que dar «vacío», no unknown_project; una
// petición explícita con un nombre inexistente sí sigue siendo error (caza erratas).
func TestJarvisProcessOverrideMayNameEmptyProject(t *testing.T) {
	s := newMCPTestStore(t)
	t.Setenv("ENGRAM_PROJECT", "")

	res, err := resolveMCPProject(s, "", "repo-solo-portatil")
	if err != nil {
		t.Fatalf("override de proceso sobre un proyecto vacío: err = %v, quiero nil", err)
	}
	if res.Project != "repo-solo-portatil" {
		t.Fatalf("proyecto = %q, quiero repo-solo-portatil", res.Project)
	}

	_, err = resolveMCPProject(s, "proyecto-con-errata", "repo-solo-portatil")
	var unknown *unknownProjectError
	if !errors.As(err, &unknown) {
		t.Fatalf("petición explícita a un proyecto inexistente: err = %v, quiero unknownProjectError", err)
	}
}

func TestJarvisRecordatorioGuardar(t *testing.T) {
	reloj := time.Unix(0, 0)
	r := nuevoRecordatorioGuardar(15*time.Minute, func() time.Time { return reloj })
	ok := r.middleware(func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	})
	llamar := func(nombre string) int {
		req := mcp.CallToolRequest{}
		req.Params.Name = nombre
		res, _ := ok(context.Background(), req)
		return len(res.Content)
	}

	if n := llamar("mem_search"); n != 1 {
		t.Fatalf("recién empezada la sesión no toca aviso: %d contenidos", n)
	}
	reloj = reloj.Add(16 * time.Minute)
	if n := llamar("mem_search"); n != 2 {
		t.Fatalf("16 min sin guardar: quiero el aviso pegado, hay %d contenidos", n)
	}
	if n := llamar("mem_context"); n != 1 {
		t.Fatalf("el aviso no se repite en cada llamada: %d contenidos", n)
	}
	reloj = reloj.Add(20 * time.Minute)
	llamar("mem_save")
	if n := llamar("mem_search"); n != 1 {
		t.Fatalf("después de guardar no toca aviso: %d contenidos", n)
	}
}
