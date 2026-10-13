package charmtest

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestKeyMsgTeclasNomeadas verifica o mapeamento das teclas nomeadas.
func TestKeyMsgTeclasNomeadas(t *testing.T) {
	casos := []struct {
		nome string
		code rune
	}{
		{"enter", tea.KeyEnter},
		{"esc", tea.KeyEscape},
		{"escape", tea.KeyEscape},
		{"tab", tea.KeyTab},
		{"backspace", tea.KeyBackspace},
		{"delete", tea.KeyDelete},
		{"up", tea.KeyUp},
		{"down", tea.KeyDown},
		{"left", tea.KeyLeft},
		{"right", tea.KeyRight},
		{"home", tea.KeyHome},
		{"end", tea.KeyEnd},
		{"pgup", tea.KeyPgUp},
		{"pgdown", tea.KeyPgDown},
		{"space", tea.KeySpace},
	}

	for _, c := range casos {
		msg, err := keyMsg(c.nome)
		if err != nil {
			t.Errorf("keyMsg(%q) retornou erro: %v", c.nome, err)
			continue
		}
		if msg.Code != c.code {
			t.Errorf("keyMsg(%q).Code = %q, esperava %q", c.nome, msg.Code, c.code)
		}
	}
}

// TestKeyMsgTeclasDeFunção verifica o suporte às teclas f1 a f12.
func TestKeyMsgTeclasDeFuncao(t *testing.T) {
	casos := []struct {
		nome string
		code rune
	}{
		{"f1", tea.KeyF1},
		{"f2", tea.KeyF2},
		{"f3", tea.KeyF3},
		{"f4", tea.KeyF4},
		{"f5", tea.KeyF5},
		{"f6", tea.KeyF6},
		{"f7", tea.KeyF7},
		{"f8", tea.KeyF8},
		{"f9", tea.KeyF9},
		{"f10", tea.KeyF10},
		{"f11", tea.KeyF11},
		{"f12", tea.KeyF12},
	}

	for _, c := range casos {
		msg, err := keyMsg(c.nome)
		if err != nil {
			t.Errorf("keyMsg(%q) retornou erro: %v", c.nome, err)
			continue
		}
		if msg.Code != c.code {
			t.Errorf("keyMsg(%q).Code = %q, esperava %q", c.nome, msg.Code, c.code)
		}
	}
}

// TestKeyMsgModificadores verifica as combinações com modificadores.
func TestKeyMsgModificadores(t *testing.T) {
	casos := []struct {
		nome string
		code rune
		mod  tea.KeyMod
	}{
		{"ctrl+c", 'c', tea.ModCtrl},
		{"ctrl+a", 'a', tea.ModCtrl},
		{"alt+x", 'x', tea.ModAlt},
		{"cmd+s", 's', tea.ModSuper},
		{"super+s", 's', tea.ModSuper},
		{"meta+s", 's', tea.ModMeta},
		{"ctrl+shift+a", 'a', tea.ModCtrl | tea.ModShift},
		{"alt+f1", tea.KeyF1, tea.ModAlt},
		{"shift+enter", tea.KeyEnter, tea.ModShift},
		{"shift+tab", tea.KeyTab, tea.ModShift},
	}

	for _, c := range casos {
		msg, err := keyMsg(c.nome)
		if err != nil {
			t.Errorf("keyMsg(%q) retornou erro: %v", c.nome, err)
			continue
		}
		if msg.Code != c.code {
			t.Errorf("keyMsg(%q).Code = %q, esperava %q", c.nome, msg.Code, c.code)
		}
		if msg.Mod != c.mod {
			t.Errorf("keyMsg(%q).Mod = %v, esperava %v", c.nome, msg.Mod, c.mod)
		}
	}
}

// TestKeyMsgShiftMaiscula é uma regressão do achado 8: em um terminal real,
// Shift+H chega como Code='h', Text="H" e Mod=ModShift.
func TestKeyMsgShiftMaiscula(t *testing.T) {
	msg, err := keyMsg("H")
	if err != nil {
		t.Fatalf("keyMsg(\"H\") retornou erro: %v", err)
	}
	if msg.Code != 'h' {
		t.Errorf("Code = %q, esperava 'h'", msg.Code)
	}
	if msg.Text != "H" {
		t.Errorf("Text = %q, esperava \"H\"", msg.Text)
	}
	if msg.Mod != tea.ModShift {
		t.Errorf("Mod = %v, esperava ModShift", msg.Mod)
	}
}

// TestKeyMsgShiftLetraMinuscula verifica shift+letra minúscula.
func TestKeyMsgShiftLetraMinuscula(t *testing.T) {
	msg, err := keyMsg("shift+a")
	if err != nil {
		t.Fatalf("keyMsg(\"shift+a\") retornou erro: %v", err)
	}
	if msg.Code != 'a' {
		t.Errorf("Code = %q, esperava 'a'", msg.Code)
	}
	if msg.Text != "A" {
		t.Errorf("Text = %q, esperava \"A\"", msg.Text)
	}
	if msg.Mod != tea.ModShift {
		t.Errorf("Mod = %v, esperava ModShift", msg.Mod)
	}
}

// TestKeyMsgMultibyte é uma regressão do achado 7: caracteres multibyte
// precisam virar teclas válidas (e não texto solto com código inválido).
func TestKeyMsgMultibyte(t *testing.T) {
	casos := []struct {
		entrada string
		code    rune
	}{
		{"é", 'é'},
		{"ç", 'ç'},
		{"漢", '漢'},
		{"😀", '😀'},
	}

	for _, c := range casos {
		msg, err := keyMsg(c.entrada)
		if err != nil {
			t.Errorf("keyMsg(%q) retornou erro: %v", c.entrada, err)
			continue
		}
		if msg.Code != c.code {
			t.Errorf("keyMsg(%q).Code = %q, esperava %q", c.entrada, msg.Code, c.code)
		}
		if msg.Text != c.entrada {
			t.Errorf("keyMsg(%q).Text = %q, esperava %q", c.entrada, msg.Text, c.entrada)
		}
		if msg.Code == 0 {
			t.Errorf("keyMsg(%q) produziu um código de tecla inválido", c.entrada)
		}
	}
}

// TestKeyMsgNomeDesconhecido é uma regressão do achado 9: um nome
// desconhecido precisa dar erro claro, e não virar texto arbitrário.
func TestKeyMsgNomeDesconhecido(t *testing.T) {
	for _, nome := range []string{"banana", "f13", "ctrl+banana", "ctrl+hyper"} {
		msg, err := keyMsg(nome)
		if err == nil {
			t.Errorf("keyMsg(%q) deveria retornar erro, obtive a mensagem %+v", nome, msg)
			continue
		}
		if !strings.Contains(err.Error(), nome) {
			t.Errorf("a mensagem de erro deveria citar %q, obtive %q", nome, err.Error())
		}
	}
}

// TestKeyMsgTeclaMais verifica que o "+" pode ser simulado.
func TestKeyMsgTeclaMais(t *testing.T) {
	msg, err := keyMsg("+")
	if err != nil {
		t.Fatalf("keyMsg(\"+\") retornou erro: %v", err)
	}
	if msg.Code != '+' || msg.Text != "+" {
		t.Errorf("keyMsg(\"+\") = %+v, esperava Code='+' e Text=\"+\"", msg)
	}

	msg, err = keyMsg("ctrl++")
	if err != nil {
		t.Fatalf("keyMsg(\"ctrl++\") retornou erro: %v", err)
	}
	if msg.Code != '+' || msg.Mod != tea.ModCtrl {
		t.Errorf("keyMsg(\"ctrl++\") = %+v, esperava Code='+' e ModCtrl", msg)
	}
}

// TestKeyMsgEspaco verifica o tratamento do espaço.
func TestKeyMsgEspaco(t *testing.T) {
	msg, err := keyMsg("space")
	if err != nil {
		t.Fatalf("keyMsg(\"space\") retornou erro: %v", err)
	}
	if msg.Code != tea.KeySpace {
		t.Errorf("Code = %q, esperava %q", msg.Code, tea.KeySpace)
	}
	if msg.Text != " " {
		t.Errorf("Text = %q, esperava um espaço", msg.Text)
	}
}

// TestKeyMsgNomeVazio verifica o erro para nome vazio.
func TestKeyMsgNomeVazio(t *testing.T) {
	if _, err := keyMsg(""); err == nil {
		t.Fatal("keyMsg(\"\") deveria retornar erro")
	}
}
