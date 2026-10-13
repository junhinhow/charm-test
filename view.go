package charmtest

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ViewSnapshot é um retrato completo da view renderizada pelo modelo.
//
// Além do conteúdo, preserva as informações que a simples string da view
// perderia: tela alternativa, cursor, cores de fundo e primeiro plano, barra
// de progresso, título da janela, modo de mouse e foco.
//
// O retrato é usado nos arquivos golden, de modo que mudanças nesses
// campos também são detectadas pelos testes.
type ViewSnapshot struct {
	// Content é o conteúdo renderizado, sem sequências ANSI.
	Content string

	// AltScreen indica se a view usa a tela alternativa.
	AltScreen bool

	// Cursor é a posição e o estilo do cursor, quando houver.
	Cursor *tea.Cursor

	// BackgroundColor é a cor de fundo do terminal, quando houver.
	BackgroundColor color.Color

	// ForegroundColor é a cor de primeiro plano do terminal, quando houver.
	ForegroundColor color.Color

	// ProgressBar é a barra de progresso do terminal, quando houver.
	ProgressBar *tea.ProgressBar

	// WindowTitle é o título da janela.
	WindowTitle string

	// ReportFocus indica se a view pede relatório de foco.
	ReportFocus bool

	// MouseMode é o modo de mouse da view.
	MouseMode tea.MouseMode

	// DisableBracketedPasteMode indica que o modo de bracketed paste está
	// desabilitado.
	DisableBracketedPasteMode bool

	// KeyboardEnhancements descreve as melhorias de teclado solicitadas.
	KeyboardEnhancements tea.KeyboardEnhancements
}

// ViewSnapshot retorna o retrato completo da view atual do modelo.
func (s *Simulator) ViewSnapshot() ViewSnapshot {
	v := s.model.View()
	return ViewSnapshot{
		Content:                   v.Content,
		AltScreen:                 v.AltScreen,
		Cursor:                    v.Cursor,
		BackgroundColor:           v.BackgroundColor,
		ForegroundColor:           v.ForegroundColor,
		ProgressBar:               v.ProgressBar,
		WindowTitle:               v.WindowTitle,
		ReportFocus:               v.ReportFocus,
		MouseMode:                 v.MouseMode,
		DisableBracketedPasteMode: v.DisableBracketedPasteMode,
		KeyboardEnhancements:      v.KeyboardEnhancements,
	}
}

// String renderiza o retrato como texto. É esse formato que vai para os
// arquivos golden.
func (vs ViewSnapshot) String() string {
	var b strings.Builder

	b.WriteString("content:\n")
	content := vs.Content
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	b.WriteString(content)

	fmt.Fprintf(&b, "altscreen: %t\n", vs.AltScreen)
	fmt.Fprintf(&b, "bracketedpaste: %t\n", !vs.DisableBracketedPasteMode)
	fmt.Fprintf(&b, "cursor: %s\n", formatCursor(vs.Cursor))
	fmt.Fprintf(&b, "background: %s\n", formatColor(vs.BackgroundColor))
	fmt.Fprintf(&b, "foreground: %s\n", formatColor(vs.ForegroundColor))
	fmt.Fprintf(&b, "mousemode: %s\n", formatMouseMode(vs.MouseMode))
	fmt.Fprintf(&b, "progressbar: %s\n", formatProgressBar(vs.ProgressBar))
	fmt.Fprintf(&b, "reportfocus: %t\n", vs.ReportFocus)
	fmt.Fprintf(&b, "windowtitle: %q\n", vs.WindowTitle)
	fmt.Fprintf(&b, "keyboardenhancements: reporteventtypes=%t\n", vs.KeyboardEnhancements.ReportEventTypes)

	return b.String()
}

// formatCursor descreve o cursor em uma linha.
func formatCursor(c *tea.Cursor) string {
	if c == nil {
		return "none"
	}
	return fmt.Sprintf("x=%d y=%d shape=%s blink=%t color=%s",
		c.X, c.Y, formatCursorShape(c.Shape), c.Blink, formatColor(c.Color))
}

// formatCursorShape nomeia a forma do cursor.
func formatCursorShape(shape tea.CursorShape) string {
	switch shape {
	case tea.CursorBlock:
		return "block"
	case tea.CursorUnderline:
		return "underline"
	case tea.CursorBar:
		return "bar"
	default:
		return "unknown"
	}
}

// formatColor converte uma cor para "#rrggbb" (ou "#rrggbbaa" quando a cor
// não for opaca).
func formatColor(c color.Color) string {
	if c == nil {
		return "none"
	}
	r, g, b, a := c.RGBA()
	if a == 0xffff {
		return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r>>8, g>>8, b>>8, a>>8)
}

// formatMouseMode nomeia o modo de mouse.
func formatMouseMode(mode tea.MouseMode) string {
	switch mode {
	case tea.MouseModeNone:
		return "none"
	case tea.MouseModeCellMotion:
		return "cellmotion"
	case tea.MouseModeAllMotion:
		return "allmotion"
	default:
		return "unknown"
	}
}

// formatProgressBar descreve a barra de progresso em uma linha.
func formatProgressBar(pb *tea.ProgressBar) string {
	if pb == nil {
		return "none"
	}
	return fmt.Sprintf("state=%s value=%d", pb.State, pb.Value)
}
