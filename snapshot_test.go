package charmtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSnapshotGoldenAusenteÉUmErro é uma regressão do achado 10: um golden
// inexistente precisa ser detectado como erro, e não passar silenciosamente.
func TestSnapshotGoldenAusenteÉUmErro(t *testing.T) {
	dir := t.TempDir()
	golden := filepath.Join(dir, "TestQualquer.golden")

	estado, err := compareSnapshot(golden, "conteúdo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if estado != snapshotMissing {
		t.Errorf("um golden inexistente deveria ser snapshotMissing, obtive %v", estado)
	}
}

// TestSnapshotGoldenIgualÉMatch verifica a comparação de um golden
// existente e igual.
func TestSnapshotGoldenIgualÉMatch(t *testing.T) {
	dir := t.TempDir()
	golden := filepath.Join(dir, "TestQualquer.golden")

	if err := os.WriteFile(golden, []byte("conteúdo"), 0o644); err != nil {
		t.Fatalf("erro ao escrever o golden: %v", err)
	}

	estado, err := compareSnapshot(golden, "conteúdo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if estado != snapshotMatch {
		t.Errorf("um golden igual deveria ser snapshotMatch, obtive %v", estado)
	}
}

// TestSnapshotGoldenDiferenteÉMismatch verifica a comparação de um golden
// existente e diferente.
func TestSnapshotGoldenDiferenteÉMismatch(t *testing.T) {
	dir := t.TempDir()
	golden := filepath.Join(dir, "TestQualquer.golden")

	if err := os.WriteFile(golden, []byte("antigo"), 0o644); err != nil {
		t.Fatalf("erro ao escrever o golden: %v", err)
	}

	estado, err := compareSnapshot(golden, "novo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if estado != snapshotMismatch {
		t.Errorf("um golden diferente deveria ser snapshotMismatch, obtive %v", estado)
	}
}

// TestSnapshotUpdateSobrescreveComConfirmacao verifica que a variável
// CHARM_TEST_UPDATE=1 é a única forma de sobrescrever um golden.
func TestSnapshotUpdateSobrescreveComConfirmacao(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("CHARM_TEST_UPDATE", "1")

	sim := New(&modeloRegistro{})
	RequireSnapshot(t, sim)

	golden := filepath.Join("testdata", "TestSnapshotUpdateSobrescreveComConfirmacao.golden")
	if _, err := os.Stat(golden); err != nil {
		t.Fatalf("o golden deveria ter sido criado: %v", err)
	}
}

// TestSnapshotCriaDiretorioDoGolden verifica que o diretório de snapshots é
// criado quando necessário.
func TestSnapshotCriaDiretorioDoGolden(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("CHARM_TEST_UPDATE", "1")

	sim := New(&modeloRegistro{})
	RequireSnapshot(t, sim)

	if _, err := os.Stat(snapshotDir); err != nil {
		t.Fatalf("o diretório de snapshots deveria ter sido criado: %v", err)
	}
}

// TestSnapshotIncluiRetratoCompleto verifica que o golden cobre também as
// informações da view que não estão no conteúdo.
func TestSnapshotIncluiRetratoCompleto(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("CHARM_TEST_UPDATE", "1")

	sim := New(&modeloCompleto{})
	RequireSnapshot(t, sim)

	golden := filepath.Join(snapshotDir, "TestSnapshotIncluiRetratoCompleto.golden")
	data, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("erro ao ler o golden: %v", err)
	}
	conteudo := string(data)

	for _, esperado := range []string{
		"content:",
		"altscreen: true",
		"cursor: x=3 y=4",
		"background: #ff0000",
		"foreground: #0000ff",
		"progressbar: state=Default value=42",
		`windowtitle: "charm-test"`,
		"reportfocus: true",
		"mousemode: allmotion",
	} {
		if !strings.Contains(conteudo, esperado) {
			t.Errorf("o golden deveria conter %q:\n%s", esperado, conteudo)
		}
	}
}

// TestRequireViewNormalizaQuebrasDeLinha é uma regressão do achado 13: a
// comparação não pode falhar por causa de \r\n.
func TestRequireViewNormalizaQuebrasDeLinha(t *testing.T) {
	sim := New(&modeloRegistro{})
	RequireView(t, sim, "80x24\r\n")
}
