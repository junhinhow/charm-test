package charmtest

import (
	"errors"
	"fmt"
	"image/color"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Cores usadas pelo modeloCompleto.
var (
	colorRed  = color.RGBA{R: 255, A: 255}
	colorBlue = color.RGBA{B: 255, A: 255}
)

// ---------------------------------------------------------------------------
// Modelos usados pelos testes do framework.
// ---------------------------------------------------------------------------

// msgA e msgB são mensagens simples usadas para verificar a ordem de
// entrega das mensagens.
type (
	msgA struct{}
	msgB struct{}
)

// modeloRegistro registra a ordem dos eventos recebidos pelo modelo.
type modeloRegistro struct {
	eventos []string
	width   int
	height  int
}

func (m *modeloRegistro) Init() tea.Cmd {
	m.eventos = append(m.eventos, "init")
	return nil
}

func (m *modeloRegistro) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.eventos = append(m.eventos, fmt.Sprintf("update:%T", msg))
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	}
	return m, nil
}

func (m *modeloRegistro) View() tea.View {
	return tea.View{Content: fmt.Sprintf("%dx%d", m.width, m.height)}
}

// modeloSequencia devolve uma tea.Sequence com dois comandos.
type modeloSequencia struct {
	recebidas []string
	rodadas   int
}

func (m *modeloSequencia) Init() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg { return msgA{} },
		func() tea.Msg { return msgB{} },
	)
}

func (m *modeloSequencia) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case msgA:
		m.recebidas = append(m.recebidas, "a")
	case msgB:
		m.recebidas = append(m.recebidas, "b")
	}
	m.rodadas++
	if m.rodadas > 2 {
		// Só devolve outra sequência por um tempo limitado: o objetivo é
		// verificar a expansão, não um loop infinito.
		return m, nil
	}
	// Devolve outra sequência, como no achado 1: sem a correção, cada
	// sequência gerava uma nova sequência, em recursão infinita.
	return m, tea.Sequence(
		func() tea.Msg { return msgA{} },
		func() tea.Msg { return msgB{} },
	)
}

func (m *modeloSequencia) View() tea.View {
	return tea.View{Content: strings.Join(m.recebidas, ",")}
}

// modeloSequenciaInfinita devolve uma sequência a cada mensagem, gerando um
// ciclo infinito de comandos.
type modeloSequenciaInfinita struct{}

func (m *modeloSequenciaInfinita) Init() tea.Cmd { return nil }

func (m *modeloSequenciaInfinita) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Sequence(
		func() tea.Msg { return msgA{} },
		func() tea.Msg { return msgB{} },
	)
}

func (m *modeloSequenciaInfinita) View() tea.View { return tea.View{} }

// modeloBatch devolve um tea.Batch com dois comandos.
type modeloBatch struct {
	recebidas []string
}

func (m *modeloBatch) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return msgA{} },
		func() tea.Msg { return msgB{} },
	)
}

func (m *modeloBatch) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case msgA:
		m.recebidas = append(m.recebidas, "a")
	case msgB:
		m.recebidas = append(m.recebidas, "b")
	}
	return m, nil
}

func (m *modeloBatch) View() tea.View {
	return tea.View{Content: strings.Join(m.recebidas, ",")}
}

// tickMsg é a mensagem de tick usada pelos modelos com temporizador.
type tickMsg struct{}

// modeloTickReal usa tea.Tick (temporizador real do Bubble Tea).
type modeloTickReal struct {
	ticks    int
	interval time.Duration
}

func (m *modeloTickReal) Init() tea.Cmd {
	return tea.Tick(m.interval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *modeloTickReal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tickMsg); ok {
		m.ticks++
		return m, tea.Tick(m.interval, func(time.Time) tea.Msg { return tickMsg{} })
	}
	return m, nil
}

func (m *modeloTickReal) View() tea.View {
	return tea.View{Content: fmt.Sprintf("ticks=%d", m.ticks)}
}

// modeloTickVirtual usa charmtest.Tick (temporizador do relógio virtual).
type modeloTickVirtual struct {
	ticks    int
	interval time.Duration
}

func (m *modeloTickVirtual) Init() tea.Cmd {
	return Tick(m.interval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *modeloTickVirtual) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tickMsg); ok {
		m.ticks++
		return m, Tick(m.interval, func(time.Time) tea.Msg { return tickMsg{} })
	}
	return m, nil
}

func (m *modeloTickVirtual) View() tea.View {
	return tea.View{Content: fmt.Sprintf("ticks=%d", m.ticks)}
}

// modeloLoop devolve sempre um comando que devolve uma mensagem, criando um
// ciclo infinito de comandos.
type modeloLoop struct{}

func (m *modeloLoop) Init() tea.Cmd { return nil }

func (m *modeloLoop) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg { return msgA{} }
}

func (m *modeloLoop) View() tea.View { return tea.View{} }

// modeloValor é um modelo com receptor por valor (o estilo mais comum).
type modeloValor struct {
	Count int
}

func (m modeloValor) Init() tea.Cmd { return nil }

func (m modeloValor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "k" {
		m.Count++
	}
	return m, nil
}

func (m modeloValor) View() tea.View {
	return tea.View{Content: fmt.Sprintf("Count: %d", m.Count)}
}

// modeloCompleto expõe todos os campos de tea.View relevantes.
type modeloCompleto struct{}

func (m *modeloCompleto) Init() tea.Cmd { return nil }

func (m *modeloCompleto) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *modeloCompleto) View() tea.View {
	v := tea.View{Content: "completo"}
	v.AltScreen = true
	v.Cursor = tea.NewCursor(3, 4)
	v.BackgroundColor = colorRed
	v.ForegroundColor = colorBlue
	v.ProgressBar = tea.NewProgressBar(tea.ProgressBarDefault, 42)
	v.WindowTitle = "charm-test"
	v.ReportFocus = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

// ---------------------------------------------------------------------------
// Testes.
// ---------------------------------------------------------------------------

// TestSequenceNaoEstouraPilha é uma regressão do achado 1: tea.Sequence
// produz um tipo não exportado que, antes, era entregue ao modelo como uma
// mensagem comum, gerando recursão infinita.
func TestSequenceNaoEstouraPilha(t *testing.T) {
	m := &modeloSequencia{}
	sim := New(m)

	if err := sim.Err(); err != nil {
		t.Fatalf("erro inesperado na simulação: %v", err)
	}
	// A sequência inicial e as duas reexpansões deliveram 3 mensagens de
	// cada tipo.
	var a, b int
	for _, r := range m.recebidas {
		switch r {
		case "a":
			a++
		case "b":
			b++
		}
	}
	if a != 3 || b != 3 {
		t.Errorf("a sequência deveria entregar 3 mensagens de cada tipo, obtive a=%d b=%d (%v)", a, b, m.recebidas)
	}
}

// TestSequenceInfinitaNaoEstouraPilha verifica que uma sequência que se
// repete indefinidamente falha com uma mensagem clara, em vez de estourar a
// pilha.
func TestSequenceInfinitaNaoEstouraPilha(t *testing.T) {
	sim := New(&modeloSequenciaInfinita{}, WithMaxDepth(20))

	err := sim.Err()
	if err == nil {
		t.Fatal("o simulador deveria ter registrado um erro de profundidade")
	}
	if !strings.Contains(err.Error(), "profundidade") {
		t.Errorf("a mensagem de erro deveria mencionar a profundidade, obtive %q", err.Error())
	}
}

// TestBatchContinuaFuncionando garante que a expansão de lotes não quebrou.
func TestBatchContinuaFuncionando(t *testing.T) {
	m := &modeloBatch{}
	New(m)

	if got := strings.Join(m.recebidas, ","); got != "a,b" {
		t.Errorf("o lote deveria entregar as mensagens a e b, obtive %q", got)
	}
}

// TestInitChamadoAntesDoWindowSizeMsg é uma regressão do achado 4: no
// Bubble Tea real, Init() é chamado antes do event loop começar, ou seja,
// antes de o WindowSizeMsg ser entregue ao Update.
func TestInitChamadoAntesDoWindowSizeMsg(t *testing.T) {
	m := &modeloRegistro{}
	New(m)

	if len(m.eventos) == 0 || m.eventos[0] != "init" {
		t.Fatalf("Init() deveria ser a primeira chamada, obtive %v", m.eventos)
	}

	esperados := []string{
		"update:tea.ColorProfileMsg",
		"update:tea.WindowSizeMsg",
		"update:tea.EnvMsg",
	}
	var updates []string
	for _, e := range m.eventos {
		if strings.HasPrefix(e, "update:") {
			updates = append(updates, e)
		}
	}
	if len(updates) != len(esperados) {
		t.Fatalf("esperava %d mensagens de inicialização, obtive %d: %v", len(esperados), len(updates), updates)
	}
	for i, esperado := range esperados {
		if updates[i] != esperado {
			t.Errorf("mensagem %d: esperava %q, obtive %q", i, esperado, updates[i])
		}
	}
}

// TestColorProfileMsgEEEnvMsgEntregues é uma regressão do achado 5.
func TestColorProfileMsgEEEnvMsgEntregues(t *testing.T) {
	m := &modeloRegistro{}
	sim := New(m, WithEnvironment([]string{"TERM=xterm-256color", "CHARM_TEST=1"}))

	var perfil *tea.ColorProfileMsg
	var env *tea.EnvMsg
	for _, msg := range sim.Messages() {
		switch msg := msg.(type) {
		case tea.ColorProfileMsg:
			cp := msg
			perfil = &cp
		case tea.EnvMsg:
			e := msg
			env = &e
		}
	}

	if perfil == nil {
		t.Fatal("o ColorProfileMsg não foi entregue ao modelo")
	}
	if env == nil {
		t.Fatal("o EnvMsg não foi entregue ao modelo")
	}
	if got := env.Getenv("CHARM_TEST"); got != "1" {
		t.Errorf("o EnvMsg deveria conter CHARM_TEST=1, obtive %q", got)
	}
	if got := env.Getenv("TERM"); got != "xterm-256color" {
		t.Errorf("o EnvMsg deveria conter TERM=xterm-256color, obtive %q", got)
	}
}

// TestTickRealNaoTrava é uma regressão do achado 2: um Init() que devolve
// tea.Tick travava a simulação para sempre.
func TestTickRealNaoTrava(t *testing.T) {
	m := &modeloTickReal{interval: 10 * time.Millisecond}

	inicio := time.Now()
	sim := New(m)
	elapsed := time.Since(inicio)

	if elapsed > 100*time.Millisecond {
		t.Fatalf("New() não deveria esperar pelo tick, levou %s", elapsed)
	}
	if err := sim.Err(); err != nil {
		t.Fatalf("erro inesperado na simulação: %v", err)
	}

	// O tick só é entregue quando o relógio virtual avança.
	if m.ticks != 0 {
		t.Errorf("nenhum tick deveria ter sido entregue antes do Advance, obtive %d", m.ticks)
	}

	sim.Advance(50 * time.Millisecond)
	if m.ticks == 0 {
		t.Error("o Advance deveria ter entregue o tick")
	}
}

// TestRelogioVirtualNaoConsomeTempoReal garante que os temporizadores
// virtuais não dormem de verdade.
func TestRelogioVirtualNaoConsomeTempoReal(t *testing.T) {
	m := &modeloTickVirtual{interval: 100 * time.Millisecond}
	sim := New(m)

	inicio := time.Now()
	sim.Advance(350 * time.Millisecond)
	elapsed := time.Since(inicio)

	if elapsed > 100*time.Millisecond {
		t.Errorf("o relógio virtual não deveria consumir tempo real, levou %s", elapsed)
	}
	if m.ticks != 3 {
		t.Errorf("esperava 3 ticks após avançar 350ms, obtive %d", m.ticks)
	}
}

// TestRelogioVirtualNowConfere o instante do relógio virtual.
func TestRelogioVirtualNowConfere(t *testing.T) {
	sim := New(&modeloRegistro{})
	if !sim.Now().Equal(epoch) {
		t.Errorf("o relógio virtual deveria começar em %s, obtive %s", epoch, sim.Now())
	}
	sim.Advance(90 * time.Second)
	if got := sim.Now().Sub(epoch); got != 90*time.Second {
		t.Errorf("esperava 90s de relógio virtual, obtive %s", got)
	}
}

// TestAdvanceUntilIdleEntregaPendencias verifica o esgotamento das
// mensagens prontas sem avanço do relógio.
func TestAdvanceUntilIdleEntregaPendencias(t *testing.T) {
	m := &modeloTickVirtual{interval: 10 * time.Millisecond}
	sim := New(m)

	sim.Advance(10 * time.Millisecond)
	antes := m.ticks
	sim.AdvanceUntilIdle()
	if m.ticks != antes {
		t.Errorf("AdvanceUntilIdle não deveria entregar ticks extras, antes=%d depois=%d", antes, m.ticks)
	}
}

// TestLimiteDeProfundidadeEvitaLoopInfinito garante que um modelo que
// devolve comandos em loop falha com uma mensagem clara, em vez de estourar
// a pilha.
func TestLimiteDeProfundidadeEvitaLoopInfinito(t *testing.T) {
	sim := New(&modeloLoop{}, WithMaxDepth(20))

	err := sim.Err()
	if err == nil {
		t.Fatal("o simulador deveria ter registrado um erro de profundidade")
	}
	if !strings.Contains(err.Error(), "profundidade") {
		t.Errorf("a mensagem de erro deveria mencionar a profundidade, obtive %q", err.Error())
	}
}

// TestLimiteDeEventosEvitaSimulacaoInfinita verifica a rede de proteção
// final contra modelos que nunca ficam ociosos.
func TestLimiteDeEventosEvitaSimulacaoInfinita(t *testing.T) {
	sim := New(&modeloLoop{}, WithMaxDepth(1000), WithMaxEvents(10))

	err := sim.Err()
	if err == nil {
		t.Fatal("o simulador deveria ter registrado um erro de limite de mensagens")
	}
	if !strings.Contains(err.Error(), "mensagens processadas") {
		t.Errorf("a mensagem de erro deveria mencionar o limite de mensagens, obtive %q", err.Error())
	}
}

// TestPendingTimersContaOsTemporizadores verifica a contagem de
// temporizadores pendentes.
func TestPendingTimersContaOsTemporizadores(t *testing.T) {
	m := &modeloTickVirtual{interval: 100 * time.Millisecond}
	sim := New(m)

	if got := sim.PendingTimers(); got != 1 {
		t.Errorf("esperava 1 temporizador pendente, obtive %d", got)
	}
	sim.Advance(100 * time.Millisecond)
	if got := sim.PendingTimers(); got != 1 {
		t.Errorf("após o tick o modelo deveria ter reagido e rearmeado o timer, obtive %d", got)
	}
}

// TestModelAsComModeloPorValor é uma regressão do achado 6: com receptor por
// valor, sim.Model().(*MeuModel) entra em pânico.
func TestModelAsComModeloPorValor(t *testing.T) {
	sim := New(modeloValor{})

	if _, ok := ModelAs[*modeloValor](sim); ok {
		t.Error("a conversão para *modeloValor não deveria funcionar em um modelo por valor")
	}

	m, ok := ModelAs[modeloValor](sim)
	if !ok {
		t.Fatal("a conversão para modeloValor deveria funcionar")
	}
	if m.Count != 0 {
		t.Errorf("esperava Count=0, obtive %d", m.Count)
	}

	if err := sim.SendKey("k"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	m, _ = ModelAs[modeloValor](sim)
	if m.Count != 1 {
		t.Errorf("esperava Count=1, obtive %d", m.Count)
	}
}

// TestMessagesRetornaCopia é uma regressão do achado 11: Messages() devolvia
// o slice interno, que podia ser modificado de fora.
func TestMessagesRetornaCopia(t *testing.T) {
	sim := New(&modeloRegistro{})

	msgs := sim.Messages()
	if len(msgs) == 0 {
		t.Fatal("esperava ao menos uma mensagem registrada")
	}
	msgs[0] = "alterado"

	if _, ok := sim.Messages()[0].(string); ok {
		t.Error("Messages() devolveu o slice interno da simulação")
	}
}

// TestLimiteDeMensagensVerifica que o histórico de mensagens é limitado.
func TestLimiteDeMensagensVerifica(t *testing.T) {
	sim := New(&modeloRegistro{}, WithMaxMessages(3))

	for i := 0; i < 10; i++ {
		sim.Send(msgA{})
	}

	if got := len(sim.Messages()); got > 3 {
		t.Errorf("o histórico deveria ter no máximo 3 mensagens, obtive %d", got)
	}
}

// TestNewComModelNuloPanica é uma regressão do achado 14.
func TestNewComModelNuloPanica(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("New() deveria entrar em pânico com um model nulo")
		}
		if !strings.Contains(fmt.Sprint(r), "model") {
			t.Errorf("a mensagem do pânico deveria mencionar o model, obtive %v", r)
		}
	}()
	New(nil)
}

// TestWithSizeInvalidoPanica é uma regressão do achado 14.
func TestWithSizeInvalidoPanica(t *testing.T) {
	for _, tamanho := range [][2]int{{0, 24}, {80, 0}, {-1, 10}} {
		func() {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("New() com tamanho %v deveria entrar em pânico", tamanho)
				}
			}()
			New(&modeloRegistro{}, WithSize(tamanho[0], tamanho[1]))
		}()
	}
}

// TestResizeValidoETamanhoInvalido verifica o redimensionamento.
func TestResizeValidoETamanhoInvalido(t *testing.T) {
	m := &modeloRegistro{}
	sim := New(m)

	sim.Resize(100, 30)
	if m.width != 100 || m.height != 30 {
		t.Errorf("esperava 100x30, obtive %dx%d", m.width, m.height)
	}

	sim.Resize(0, 10)
	if sim.Err() == nil {
		t.Error("um resize inválido deveria registrar um erro")
	}
}

// TestViewSnapshotPreservaCampos é uma regressão do achado 12: a view
// precisa expor mais do que o conteúdo.
func TestViewSnapshotPreservaCampos(t *testing.T) {
	sim := New(&modeloCompleto{})
	snapshot := sim.ViewSnapshot()

	if !snapshot.AltScreen {
		t.Error("o retrato deveria preservar o AltScreen")
	}
	if snapshot.Cursor == nil || snapshot.Cursor.X != 3 || snapshot.Cursor.Y != 4 {
		t.Errorf("o retrato deveria preservar o cursor, obtive %+v", snapshot.Cursor)
	}
	if snapshot.BackgroundColor == nil {
		t.Error("o retrato deveria preservar a cor de fundo")
	}
	if snapshot.ProgressBar == nil || snapshot.ProgressBar.Value != 42 {
		t.Errorf("o retrato deveria preservar a barra de progresso, obtive %+v", snapshot.ProgressBar)
	}
	if snapshot.WindowTitle != "charm-test" {
		t.Errorf("o retrato deveria preservar o título, obtive %q", snapshot.WindowTitle)
	}
	if !snapshot.ReportFocus {
		t.Error("o retrato deveria preservar o ReportFocus")
	}
	if snapshot.MouseMode != tea.MouseModeAllMotion {
		t.Errorf("o retrato deveria preservar o modo de mouse, obtive %v", snapshot.MouseMode)
	}
}

// TestFullViewExpoeViewCompleta verifica o acesso à tea.View completa.
func TestFullViewExpoeViewCompleta(t *testing.T) {
	sim := New(&modeloCompleto{})
	v := sim.FullView()

	if !v.AltScreen || v.WindowTitle != "charm-test" {
		t.Errorf("FullView deveria devolver a view completa, obtive %+v", v)
	}
}

// TestSendKeyDesconhecidoRetornaErro é uma regressão do achado 9: um nome
// desconhecido não pode virar texto arbitrário entregue ao modelo.
func TestSendKeyDesconhecidoRetornaErro(t *testing.T) {
	sim := New(&modeloRegistro{})

	err := sim.SendKey("banana")
	if err == nil {
		t.Fatal("SendKey deveria retornar um erro para um nome desconhecido")
	}
	if !strings.Contains(err.Error(), "banana") {
		t.Errorf("a mensagem de erro deveria citar o nome inválido, obtive %q", err.Error())
	}
	if sim.Err() == nil {
		t.Error("o simulador deveria ter registrado o erro")
	}

	antes := len(sim.Messages())
	if err := sim.SendKey("f13"); err == nil {
		t.Error("SendKey deveria retornar um erro para f13")
	}
	if len(sim.Messages()) != antes {
		t.Error("nenhuma mensagem deveria ter sido entregue ao modelo")
	}
}

// TestTypeEnviaCadaCaractere verifica a digitação de texto.
func TestTypeEnviaCadaCaractere(t *testing.T) {
	m := &modeloValor{}
	sim := New(m)

	sim.Type("abc")
	got, _ := ModelAs[modeloValor](sim)
	if got.Count != 0 {
		t.Errorf("as letras comuns não deveriam alterar a contagem, obtive %d", got.Count)
	}

	sim.Type("k")
	got, _ = ModelAs[modeloValor](sim)
	if got.Count != 1 {
		t.Errorf("esperava Count=1, obtive %d", got.Count)
	}
}

// TestViewContainsVerifica a busca de substring na view.
func TestViewContainsVerifica(t *testing.T) {
	sim := New(&modeloValor{})
	if !sim.ViewContains("Count: 0") {
		t.Errorf("a view deveria conter %q, obtive %q", "Count: 0", sim.View())
	}
	if sim.ViewContains("inexistente") {
		t.Error("a view não deveria conter texto inexistente")
	}
}

// TestErrRetornaErroRegistrado verifica o accessor de erro.
func TestErrRetornaErroRegistrado(t *testing.T) {
	sim := New(&modeloRegistro{})
	if sim.Err() != nil {
		t.Errorf("o simulador não deveria ter erro, obtive %v", sim.Err())
	}

	sim.Resize(-1, -1)
	if sim.Err() == nil {
		t.Fatal("o simulador deveria ter registrado o erro")
	}
	if !errors.Is(sim.Err(), sim.Err()) {
		t.Error("o accessor de erro deveria ser estável")
	}
}
