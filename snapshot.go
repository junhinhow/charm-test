package charmtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

const snapshotDir = "testdata"

// RequireSnapshot compara a view atual com um arquivo golden armazenado em
// testdata/<nome-do-teste>.golden.
//
// O golden precisa existir: se ele não existir, o teste falha com uma
// mensagem explicando como criá-lo. Isso evita que uma mudança de output
// passe despercebida em CI, em que todo snapshot novo seria aceito sem
// comparação.
//
// Use a variável de ambiente CHARM_TEST_UPDATE=1 para criar ou sobrescrever
// os arquivos golden:
//
//	CHARM_TEST_UPDATE=1 go test ./...
//
// O snapshot inclui o retrato completo da view (conteúdo, tela alternativa,
// cursor, cores, barra de progresso e título), e não apenas o conteúdo.
//
// Uso:
//
//	func TestMyView(t *testing.T) {
//	    sim := charmtest.New(NewMyModel())
//	    sim.SendKey("j")
//	    charmtest.RequireSnapshot(t, sim)
//	}
func RequireSnapshot(t *testing.T, s *Simulator) {
	t.Helper()
	requireSnapshotNamed(t, s, t.Name())
}

// RequireSnapshotNamed é como RequireSnapshot, mas usa um nome customizado
// para o arquivo golden. Útil quando um único teste captura vários snapshots
// em pontos diferentes.
func RequireSnapshotNamed(t *testing.T, s *Simulator, name string) {
	t.Helper()
	requireSnapshotNamed(t, s, name)
}

func requireSnapshotNamed(t *testing.T, s *Simulator, name string) {
	t.Helper()
	requireNoError(t, s)

	// Sanitiza o nome para uso como nome de arquivo.
	safeName := strings.ReplaceAll(name, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	safeName = strings.ReplaceAll(safeName, " ", "_")

	goldenPath := filepath.Join(snapshotDir, safeName+".golden")
	actual := normalizeNewlines(s.ViewSnapshot().String())
	actual = ansi.Strip(actual)

	update := os.Getenv("CHARM_TEST_UPDATE") == "1"

	estado, err := compareSnapshot(goldenPath, actual)
	switch {
	case err != nil:
		t.Fatalf("falha ao ler o snapshot golden %s: %v", goldenPath, err)
	case estado == snapshotMatch:
		return
	case estado == snapshotMissing:
		if !update {
			t.Fatalf("snapshot golden não encontrado: %s\nCrie-o com: CHARM_TEST_UPDATE=1 go test ./...", goldenPath)
		}
		writeGolden(t, goldenPath, actual)
		t.Logf("snapshot criado: %s", goldenPath)
	case estado == snapshotMismatch:
		expected, _ := os.ReadFile(goldenPath)
		if !update {
			t.Errorf("snapshot divergente em %s:\n--- esperado (golden) ---\n%s\n--- atual ---\n%s\n\nAtualize com: CHARM_TEST_UPDATE=1 go test ./...",
				goldenPath, normalizeNewlines(string(expected)), actual)
			return
		}
		writeGolden(t, goldenPath, actual)
		t.Logf("snapshot atualizado: %s", goldenPath)
	}
}

// snapshotEstado é o resultado da comparação entre o snapshot atual e o
// arquivo golden.
type snapshotEstado int

const (
	// snapshotMatch indica que o golden existe e é igual ao snapshot atual.
	snapshotMatch snapshotEstado = iota
	// snapshotMismatch indica que o golden existe, mas difere do snapshot.
	snapshotMismatch
	// snapshotMissing indica que o golden ainda não existe.
	snapshotMissing
)

// compareSnapshot compara o snapshot atual com o arquivo golden.
//
// As quebras de linha são normalizadas, para que o mesmo golden funcione em
// Windows e em sistemas Unix, independentemente da configuração do Git.
func compareSnapshot(goldenPath, actual string) (snapshotEstado, error) {
	expected, err := os.ReadFile(goldenPath)
	switch {
	case err == nil:
		if normalizeNewlines(string(expected)) == actual {
			return snapshotMatch, nil
		}
		return snapshotMismatch, nil
	case os.IsNotExist(err):
		return snapshotMissing, nil
	default:
		return snapshotMatch, err
	}
}

// normalizeNewlines converte as quebras de linha CRLF e CR em LF.
func normalizeNewlines(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

// writeGolden grava o conteúdo do snapshot no arquivo golden.
func writeGolden(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf("falha ao criar o diretório de snapshots: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("falha ao escrever o snapshot golden: %v", err)
	}
}
