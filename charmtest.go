// Package charmtest provides a testing framework for Bubble Tea applications.
//
// O pacote simula o runtime do Bubble Tea sem precisar de um terminal real,
// permitindo enviar mensagens, simular teclas, avançar um relógio virtual e
// verificar o estado do modelo e a saída renderizada.
//
// Uso básico:
//
//	func TestMyModel(t *testing.T) {
//	    m := NewMyModel()
//	    sim := charmtest.New(m)
//	    sim.SendKey("j")           // simula a tecla "j"
//	    sim.SendKey("enter")       // simula a tecla enter
//	    charmtest.RequireView(t, sim, "expected output")
//	}
//
// # Relógio virtual
//
// Apps com spinner, relógio ou qualquer temporizador usam tea.Tick ou
// tea.Every. Esses comandos bloqueiam até o prazo, o que tornaria a simulação
// infinita. Aqui eles são interceptados e adiados: a mensagem só é entregue
// quando o relógio virtual avança, via [Simulator.Advance].
//
// Para uma simulação 100% virtual (sem qualquer consumo de tempo real),
// use [Tick] e [Every] no lugar de tea.Tick e tea.Every.
package charmtest

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

const (
	// DefaultWidth é a largura padrão do terminal simulado.
	DefaultWidth = 80

	// DefaultHeight é a altura padrão do terminal simulado.
	DefaultHeight = 24

	// DefaultMaxMessages é a quantidade máxima de mensagens guardadas pelo
	// simulador. As mensagens mais antigas são descartadas quando o limite
	// é atingido.
	DefaultMaxMessages = 1000

	// DefaultMaxDepth é a profundidade máxima de processamento de comandos.
	// Ela protege a simulação contra modelos que devolvem comandos em
	// loop (por exemplo, um tick que se rearma indefinidamente).
	DefaultMaxDepth = 100

	// DefaultMaxEvents é o número máximo de mensagens processadas durante
	// toda a simulação. É a última rede de proteção contra modelos que
	// nunca ficam ociosos.
	DefaultMaxEvents = 100000
)

// Simulator conduz um modelo Bubble Tea pelo seu ciclo de vida sem um
// terminal real. As mensagens são processadas de forma síncrona, o que torna
// os testes determinísticos.
type Simulator struct {
	model  tea.Model
	width  int
	height int

	// messages guarda todas as mensagens processadas, para depuração.
	messages    []tea.Msg
	maxMessages int

	// depth controla a profundidade do processamento de comandos e events o
	// total de mensagens processadas, evitando loops infinitos.
	depth     int
	maxDepth  int
	events    int
	maxEvents int

	// now é o instante atual do relógio virtual e scheduleBase é a base
	// usada para calcular o prazo dos novos temporizadores. Durante a
	// reação a um timer atrasado, scheduleBase é o prazo do timer, para que
	// o próximo ciclo do modelo continue alinhado à grade original.
	now          time.Time
	scheduleBase time.Time

	// timers são os temporizadores virtuais pendentes; realTimers são os
	// temporizadores reais (tea.Tick/tea.Every) já em execução.
	timers     []*virtualTimer
	realTimers []*realTimer

	// profile e environ são entregues ao modelo em ColorProfileMsg e EnvMsg,
	// como o Bubble Tea real faz na inicialização.
	profile colorprofile.Profile
	environ []string

	// err guarda o primeiro erro encontrado durante a simulação.
	err error
}

// Option configura um Simulator.
type Option func(*Simulator)

// WithSize define as dimensões do terminal simulado. As dimensões precisam ser
// maiores que zero, caso contrário New entra em pânico com uma mensagem
// explicativa.
func WithSize(width, height int) Option {
	return func(s *Simulator) {
		s.width = width
		s.height = height
	}
}

// WithMaxMessages limita quantas mensagens o simulador guarda para
// depuração. As mais antigas são descartadas quando o limite é atingido.
func WithMaxMessages(n int) Option {
	return func(s *Simulator) {
		s.maxMessages = n
	}
}

// WithMaxDepth limita a profundidade do processamento de comandos. Serve
// como rede de proteção contra modelos que devolvem comandos em loop.
func WithMaxDepth(n int) Option {
	return func(s *Simulator) {
		s.maxDepth = n
	}
}

// WithMaxEvents limita o total de mensagens processadas na simulação. É a
// rede de proteção final contra modelos que nunca ficam ociosos.
func WithMaxEvents(n int) Option {
	return func(s *Simulator) {
		s.maxEvents = n
	}
}

// WithColorProfile define o perfil de cores entregue ao modelo em
// tea.ColorProfileMsg. O padrão é colorprofile.TrueColor.
func WithColorProfile(profile colorprofile.Profile) Option {
	return func(s *Simulator) {
		s.profile = profile
	}
}

// WithEnvironment define as variáveis de ambiente entregues ao modelo em
// tea.EnvMsg. O padrão é o ambiente do processo (os.Environ).
func WithEnvironment(environ []string) Option {
	return func(s *Simulator) {
		s.environ = environ
	}
}

// New cria um Simulator que envolve o modelo Bubble Tea informado.
//
// A inicialização segue a mesma ordem do Bubble Tea real:
//
//  1. Init() é chamado (no Bubble Tea real ele é chamado antes de o event
//     loop começar, portanto antes de qualquer mensagem);
//  2. tea.ColorProfileMsg, tea.WindowSizeMsg e tea.EnvMsg são entregues ao
//     modelo, nessa ordem;
//  3. o comando devolvido por Init() é processado.
//
// New entra em pânico se o model for nulo ou se as dimensões do terminal
// forem inválidas (veja [WithSize]).
func New(model tea.Model, opts ...Option) *Simulator {
	if model == nil {
		panic("charmtest: New requer um model não nulo")
	}

	s := &Simulator{
		model:       model,
		width:       DefaultWidth,
		height:      DefaultHeight,
		maxMessages: DefaultMaxMessages,
		maxDepth:    DefaultMaxDepth,
		maxEvents:   DefaultMaxEvents,
		profile:     colorprofile.TrueColor,
		environ:     os.Environ(),
		now:         epoch,
	}
	s.scheduleBase = epoch
	for _, opt := range opts {
		opt(s)
	}

	if s.width <= 0 || s.height <= 0 {
		panic(fmt.Sprintf("charmtest: dimensões do terminal inválidas: %dx%d (devem ser maiores que zero)", s.width, s.height))
	}
	if s.maxMessages <= 0 {
		s.maxMessages = DefaultMaxMessages
	}
	if s.maxDepth <= 0 {
		s.maxDepth = DefaultMaxDepth
	}
	if s.maxEvents <= 0 {
		s.maxEvents = DefaultMaxEvents
	}
	if s.environ == nil {
		s.environ = []string{}
	}

	// No Bubble Tea real, Init() é chamado antes do event loop começar,
	// ou seja, antes de WindowSizeMsg (e das demais mensagens de
	// inicialização) serem entregues ao Update.
	initCmd := s.model.Init()

	// O Bubble Tea real envia ColorProfileMsg, WindowSizeMsg e EnvMsg
	// nessa ordem quando o programa inicia.
	s.Send(tea.ColorProfileMsg{Profile: s.profile})
	s.Send(tea.WindowSizeMsg{Width: s.width, Height: s.height})
	s.Send(tea.EnvMsg(s.environ))

	// Só então o comando devolvido por Init() é processado.
	s.processCmd(initCmd)

	return s
}

// Send envia uma mensagem ao modelo e processa o comando resultante.
func (s *Simulator) Send(msg tea.Msg) {
	if s.events >= s.maxEvents {
		s.setErr(fmt.Errorf("charmtest: limite de %d mensagens processadas excedido", s.maxEvents))
		return
	}
	s.events++

	s.record(msg)

	var cmd tea.Cmd
	s.model, cmd = s.model.Update(msg)
	s.processCmd(cmd)
}

// SendKey simula o pressionar de uma tecla. Aceita nomes de teclas
// ("enter", "esc", "tab", "f1", "shift+enter", "ctrl+c", ...) ou um único
// caractere ("a", "j", "é").
//
// Nomes desconhecidos não viram texto arbitrário entregue ao modelo: um erro
// é retornado e registrado no simulador (veja [Simulator.Err]), o que faz as
// assertions falharem com uma mensagem clara.
func (s *Simulator) SendKey(key string) error {
	msg, err := keyMsg(key)
	if err != nil {
		s.setErr(err)
		return err
	}
	s.Send(msg)
	return nil
}

// Type simula a digitação de um texto, enviando uma tecla por caractere.
func (s *Simulator) Type(text string) {
	for _, r := range text {
		msg, err := keyMsg(string(r))
		if err != nil {
			s.setErr(err)
			continue
		}
		s.Send(msg)
	}
}

// SendKeys simula uma sequência de teclas.
func (s *Simulator) SendKeys(keys ...string) error {
	for _, k := range keys {
		if err := s.SendKey(k); err != nil {
			return err
		}
	}
	return nil
}

// View retorna o conteúdo renderizado pelo modelo.
func (s *Simulator) View() string {
	return s.model.View().Content
}

// FullView retorna a tea.View completa renderizada pelo modelo, incluindo
// informações que não estão no conteúdo (tela alternativa, cursor, cores,
// barra de progresso e título da janela).
func (s *Simulator) FullView() tea.View {
	return s.model.View()
}

// Model retorna o estado atual do modelo.
//
// Modelos com receptor por valor não podem ser acessados por type assertion
// direto (o tipo guardado é o valor, não o ponteiro). Nesse caso use
// [ModelAs].
func (s *Simulator) Model() tea.Model {
	return s.model
}

// Messages retorna uma cópia de todas as mensagens enviadas ao modelo.
// Modificar o resultado não afeta o estado interno do simulador.
func (s *Simulator) Messages() []tea.Msg {
	out := make([]tea.Msg, len(s.messages))
	copy(out, s.messages)
	return out
}

// Err retorna o primeiro erro registrado durante a simulação, se houver.
// As assertions do pacote falham automaticamente quando há erro.
func (s *Simulator) Err() error {
	return s.err
}

// Resize simula um evento de redimensionamento do terminal.
func (s *Simulator) Resize(width, height int) {
	if width <= 0 || height <= 0 {
		s.setErr(fmt.Errorf("charmtest: dimensões do terminal inválidas: %dx%d (devem ser maiores que zero)", width, height))
		return
	}
	s.width = width
	s.height = height
	s.Send(tea.WindowSizeMsg{Width: width, Height: height})
}

// ViewContains informa se a view atual contém a substring informada.
func (s *Simulator) ViewContains(substr string) bool {
	return strings.Contains(s.View(), substr)
}

// record guarda a mensagem para depuração, respeitando o limite configurado.
func (s *Simulator) record(msg tea.Msg) {
	s.messages = append(s.messages, msg)
	if len(s.messages) > s.maxMessages {
		copy(s.messages, s.messages[len(s.messages)-s.maxMessages:])
		s.messages = s.messages[:s.maxMessages]
	}
}

// setErr registra o primeiro erro da simulação.
func (s *Simulator) setErr(err error) {
	if err == nil || s.err != nil {
		return
	}
	s.err = err
}

// processCmd executa um comando e entrega a mensagem resultante ao modelo.
//
// Lotes (tea.Batch) e sequências (tea.Sequence) são expandidos. Os
// temporizadores do Bubble Tea (tea.Tick e tea.Every) são adiados, porque
// bloqueiam até o prazo e fariam a simulação nunca terminar.
func (s *Simulator) processCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	if s.depth >= s.maxDepth {
		s.setErr(fmt.Errorf(
			"charmtest: limite de profundidade de comandos (%d) excedido; o modelo provavelmente devolve comandos em loop (use Advance para controlar o tempo)",
			s.maxDepth,
		))
		return
	}
	s.depth++
	defer func() { s.depth-- }()

	if isBlockingTimer(cmd) {
		s.scheduleRealTimer(cmd)
		return
	}

	msg := cmd()
	if msg == nil {
		return
	}
	s.handleMsg(msg)
}

// handleMsg entrega a mensagem ao modelo, expandindo lotes e sequências.
func (s *Simulator) handleMsg(msg tea.Msg) {
	switch m := msg.(type) {
	case *virtualTimer:
		s.scheduleVirtualTimer(m)
		return
	case tea.BatchMsg:
		for _, c := range m {
			s.processCmd(c)
		}
		return
	case tea.QuitMsg:
		// Mensagens de encerramento são ignoradas nos testes.
		return
	}

	if cmds, ok := sequenceCmds(msg); ok {
		for _, c := range cmds {
			s.processCmd(c)
		}
		return
	}

	s.Send(msg)
}

// ModelAs converte o modelo atual do simulador para o tipo T.
//
// Ele resolve o problema dos modelos com receptor por valor, em que
// sim.Model().(*MeuModel) entra em pânico porque o tipo guardado é o valor
// (e não o ponteiro).
//
// Exemplo:
//
//	m, ok := charmtest.ModelAs[MeuModel](sim)
//	if !ok {
//	    t.Fatal("modelo inesperado")
//	}
func ModelAs[T tea.Model](s *Simulator) (T, bool) {
	var zero T
	if s == nil || s.model == nil {
		return zero, false
	}
	m, ok := s.model.(T)
	return m, ok
}
