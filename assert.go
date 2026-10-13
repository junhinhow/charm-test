package charmtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// RequireView verifica se a view atual é exatamente igual à string esperada.
// As sequências ANSI são removidas antes da comparação e as quebras de linha
// Windows (\r\n) são normalizadas, para que o mesmo teste funcione em todos
// os sistemas operacionais.
func RequireView(t *testing.T, s *Simulator, expected string) {
	t.Helper()
	requireNoError(t, s)
	actual := normalizeView(s.View())
	expected = normalizeView(expected)
	if actual != expected {
		t.Errorf("view divergente:\n--- esperado ---\n%s\n--- atual ---\n%s", expected, actual)
	}
}

// RequireViewContains verifica se a view (sem ANSI) contém a substring
// informada.
func RequireViewContains(t *testing.T, s *Simulator, substr string) {
	t.Helper()
	requireNoError(t, s)
	actual := ansi.Strip(s.View())
	if !strings.Contains(actual, substr) {
		t.Errorf("a view não contém %q:\n%s", substr, actual)
	}
}

// RequireViewNotContains verifica se a view (sem ANSI) NÃO contém a
// substring informada.
func RequireViewNotContains(t *testing.T, s *Simulator, substr string) {
	t.Helper()
	requireNoError(t, s)
	actual := ansi.Strip(s.View())
	if strings.Contains(actual, substr) {
		t.Errorf("a view não deveria conter %q, mas contém:\n%s", substr, actual)
	}
}

// RequireViewLines verifica se a view tem exatamente a quantidade esperada de
// linhas não vazias.
func RequireViewLines(t *testing.T, s *Simulator, expected int) {
	t.Helper()
	requireNoError(t, s)
	actual := ansi.Strip(s.View())
	lines := 0
	for _, line := range strings.Split(actual, "\n") {
		if strings.TrimSpace(line) != "" {
			lines++
		}
	}
	if lines != expected {
		t.Errorf("esperava %d linhas não vazias, obtive %d", expected, lines)
	}
}

// ViewString retorna a view atual sem as sequências ANSI, útil para
// depuração nos testes.
func ViewString(s *Simulator) string {
	return ansi.Strip(s.View())
}

// DumpView imprime a view atual (sem ANSI) no log do teste, para
// depuração.
func DumpView(t *testing.T, s *Simulator) {
	t.Helper()
	t.Logf("View atual:\n%s", ViewString(s))
}

// DumpMessages imprime todas as mensagens processadas pelo simulador no log
// do teste, para depuração.
func DumpMessages(t *testing.T, s *Simulator) {
	t.Helper()
	var b strings.Builder
	for i, msg := range s.Messages() {
		fmt.Fprintf(&b, "  [%d] %T: %v\n", i, msg, msg)
	}
	t.Logf("Mensagens (%d):\n%s", len(s.Messages()), b.String())
}

// requireNoError falha o teste quando o simulador registrou algum erro
// (por exemplo, um nome de tecla desconhecido ou um loop de comandos).
func requireNoError(t *testing.T, s *Simulator) {
	t.Helper()
	if s == nil {
		t.Fatal("charmtest: simulador nulo")
		return
	}
	if err := s.Err(); err != nil {
		t.Fatalf("erro na simulação: %v", err)
	}
}

// normalizeView normaliza uma view para comparação: remove ANSI, converte
// \r\n em \n e apaga espaços nas pontas.
func normalizeView(view string) string {
	return strings.TrimSpace(strings.ReplaceAll(ansi.Strip(view), "\r\n", "\n"))
}
