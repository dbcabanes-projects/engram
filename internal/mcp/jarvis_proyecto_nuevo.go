package mcp

import (
	"fmt"

	"github.com/Gentleman-Programming/engram/v3/internal/store"
)

// Copia de JARVIS: el portátil abre el MCP con `--project <carpeta>`. Si esa carpeta aún no tiene memorias,
// el primer mem_save crearía el proyecto con ese nombre sin que nadie lo decida. Las instrucciones del
// servidor avisan al agente para que lo confirme con Dani si no está claro.
func avisoProyectoNuevo(s *store.Store, proyecto string) string {
	if proyecto == "" {
		return ""
	}
	if existe, err := s.ProjectExists(proyecto); err != nil || existe {
		return ""
	}
	return fmt.Sprintf("\n\nNEW PROJECT: this session's Engram project is %q (taken from the folder the agent was opened in) "+
		"and it has no memories yet. Before the first mem_save, call mem_list_projects. If this work clearly deserves "+
		"its own project, just save (that creates %q). If it may belong to an existing project, or the name is not "+
		"obvious, ask the user first; to save into an existing project pass project=<its name> in mem_save.", proyecto, proyecto)
}
