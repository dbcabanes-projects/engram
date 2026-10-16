package mcp

import (
	"context"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Copia de JARVIS: «que guarde solo por defecto» sin hooks en el cliente.
//
// El plugin oficial recuerda guardar desde un hook de la máquina del agente (UserPromptSubmit, cada 15
// min sin guardar). El portátil entra por SSH sin plugins, así que el recordatorio lo pone el servidor:
// si este proceso MCP —una sesión de agente— lleva `cada` sin una escritura, la siguiente respuesta
// correcta de cualquier herramienta lleva el aviso pegado. Límite sabido: el agente solo lo ve cuando
// llama a una herramienta de Engram; lo que no consulta la memoria no se entera.
// ponytail: un reloj por proceso, no por sesión de Engram; vale porque cada SSH es una sesión de agente.

const avisoGuardar = "MEMORY REMINDER: no memory has been saved in this session for a while. " +
	"If you have made decisions, fixed bugs, discovered something or finished significant work, call mem_save now " +
	"(and mem_session_summary before ending). Save proactively; the user should not have to ask."

var herramientasQueGuardan = map[string]bool{
	"mem_save":             true,
	"mem_update":           true,
	"mem_session_summary":  true,
	"mem_session_finalize": true,
	"mem_capture_passive":  true,
	"mem_save_prompt":      false, // guardar el prompt del usuario no es guardar lo aprendido
}

type recordatorioGuardar struct {
	mu     sync.Mutex
	ultimo time.Time
	cada   time.Duration
	ahora  func() time.Time
}

func nuevoRecordatorioGuardar(cada time.Duration, ahora func() time.Time) *recordatorioGuardar {
	return &recordatorioGuardar{ultimo: ahora(), cada: cada, ahora: ahora}
}

func (r *recordatorioGuardar) middleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := next(ctx, req)
		if err != nil || res == nil || res.IsError {
			return res, err
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		if herramientasQueGuardan[req.Params.Name] {
			r.ultimo = r.ahora()
			return res, err
		}
		if r.ahora().Sub(r.ultimo) >= r.cada {
			res.Content = append(res.Content, mcp.NewTextContent(avisoGuardar))
			r.ultimo = r.ahora() // un aviso por tramo, no en cada llamada
		}
		return res, err
	}
}
