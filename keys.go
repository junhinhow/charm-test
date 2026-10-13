package charmtest

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// namedKeys mapeia nomes de teclas para o código de tecla correspondente.
var namedKeys = map[string]rune{
	"enter":       tea.KeyEnter,
	"return":      tea.KeyReturn,
	"esc":         tea.KeyEscape,
	"escape":      tea.KeyEscape,
	"tab":         tea.KeyTab,
	"backspace":   tea.KeyBackspace,
	"delete":      tea.KeyDelete,
	"del":         tea.KeyDelete,
	"insert":      tea.KeyInsert,
	"up":          tea.KeyUp,
	"down":        tea.KeyDown,
	"left":        tea.KeyLeft,
	"right":       tea.KeyRight,
	"home":        tea.KeyHome,
	"end":         tea.KeyEnd,
	"pgup":        tea.KeyPgUp,
	"pageup":      tea.KeyPgUp,
	"pgdown":      tea.KeyPgDown,
	"pagedown":    tea.KeyPgDown,
	"space":       tea.KeySpace,
	"f1":          tea.KeyF1,
	"f2":          tea.KeyF2,
	"f3":          tea.KeyF3,
	"f4":          tea.KeyF4,
	"f5":          tea.KeyF5,
	"f6":          tea.KeyF6,
	"f7":          tea.KeyF7,
	"f8":          tea.KeyF8,
	"f9":          tea.KeyF9,
	"f10":         tea.KeyF10,
	"f11":         tea.KeyF11,
	"f12":         tea.KeyF12,
	"capslock":    tea.KeyCapsLock,
	"scrolllock":  tea.KeyScrollLock,
	"numlock":     tea.KeyNumLock,
	"printscreen": tea.KeyPrintScreen,
	"pause":       tea.KeyPause,
	"menu":        tea.KeyMenu,
}

// modifiers mapeia nomes de modificadores para o valor correspondente.
var modifiers = map[string]tea.KeyMod{
	"ctrl":  tea.ModCtrl,
	"alt":   tea.ModAlt,
	"shift": tea.ModShift,
	"cmd":   tea.ModSuper,
	"super": tea.ModSuper,
	"meta":  tea.ModMeta,
	"hyper": tea.ModHyper,
}

// keyMsg converte um nome de tecla em um tea.KeyPressMsg.
//
// São aceitos:
//
//   - teclas nomeadas ("enter", "esc", "tab", "f1"..."f12", "up", ...);
//   - combinações de modificadores ("ctrl+c", "alt+f1", "shift+enter",
//     "cmd+s", "ctrl+shift+a");
//   - um único caractere ("a", "j", "é", "漢").
//
// Teclas maiúsculas seguem o comportamento de um terminal real: "H" vira
// Code='h', Text="H" e Mod=ModShift.
//
// Nomes desconhecidos geram um erro: nenhum texto arbitrário é entregue ao
// modelo.
func keyMsg(key string) (tea.KeyPressMsg, error) {
	if key == "" {
		return tea.KeyPressMsg{}, fmt.Errorf("charmtest: nome de tecla vazio")
	}

	mods, name := splitKey(key)

	var mod tea.KeyMod
	for _, m := range mods {
		value, ok := modifiers[strings.ToLower(m)]
		if !ok {
			return tea.KeyPressMsg{}, fmt.Errorf(
				"charmtest: modificador de tecla desconhecido: %q (em %q)", m, key,
			)
		}
		mod |= value
	}

	// Tecla nomeada.
	if code, ok := namedKeys[strings.ToLower(name)]; ok {
		msg := tea.KeyPressMsg{Code: code, Mod: mod}
		if code == tea.KeySpace {
			// O espaço é o único caractere imprimível invisível.
			msg.Text = " "
		}
		return msg, nil
	}

	// Caractere único.
	runes := []rune(name)
	if len(runes) != 1 {
		return tea.KeyPressMsg{}, fmt.Errorf(
			"charmtest: nome de tecla desconhecido: %q (use um nome de tecla, uma combinação com modificador ou um único caractere)", key,
		)
	}
	r := runes[0]

	msg := tea.KeyPressMsg{Code: r, Mod: mod}
	// Teclas de controle (ctrl, alt, cmd, ...) não produzem texto
	// imprimível, como em um terminal real.
	if mod&^(tea.ModShift) != 0 {
		return msg, nil
	}

	switch {
	case unicode.IsUpper(r) && r < utf8.RuneSelf:
		// Letra maiúscula: o código é a letra minúscula e o texto é a
		// maiúscula, exatamente como o terminal entrega (shift).
		msg.Code = unicode.ToLower(r)
		msg.Text = string(r)
		msg.Mod |= tea.ModShift
	case mod&tea.ModShift != 0 && r < utf8.RuneSelf:
		// shift+letra minúscula: o texto é a versão maiúscula.
		if upper := unicode.ToUpper(r); upper != r {
			msg.Code = unicode.ToLower(r)
			msg.Text = string(upper)
			return msg, nil
		}
		msg.Text = string(r)
	default:
		msg.Text = string(r)
	}
	return msg, nil
}

// splitKey separa os modificadores do nome da tecla. O "+" é aceito tanto
// como separador quanto como caractere ("ctrl++" é ctrl+"+").
func splitKey(key string) (mods []string, name string) {
	parts := strings.Split(key, "+")
	if len(parts) == 1 {
		return nil, parts[0]
	}
	if parts[len(parts)-1] == "" {
		// A tecla é o próprio "+": "ctrl++".
		mods = parts[:len(parts)-1]
		if len(mods) > 0 && mods[len(mods)-1] == "" {
			mods = mods[:len(mods)-1]
		}
		return mods, "+"
	}
	return parts[:len(parts)-1], parts[len(parts)-1]
}
