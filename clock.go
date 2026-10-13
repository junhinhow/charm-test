package charmtest

import (
	"fmt"
	"reflect"
	"time"

	tea "charm.land/bubbletea/v2"
)

// epoch é o instante inicial do relógio virtual. É fixo para que a simulação
// seja reprodutível: o mesmo conjunto de entradas produz sempre as mesmas
// saídas, independentemente do relógio da máquina.
var epoch = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

// virtualTimer é um temporizador do relógio virtual, criado por [Tick] ou
// [Every]. Ele só dispara quando o relógio virtual avança.
type virtualTimer struct {
	dur   time.Duration
	every bool
	fn    func(time.Time) tea.Msg
	due   time.Time
}

// realTimer é um temporizador real (criado por tea.Tick ou tea.Every) que já
// está rodando em uma goroutine. O Bubble Tea entrega a mensagem do
// temporizador apenas quando o timer real dispara, portanto o simulador
// precisa esperar por ela.
type realTimer struct {
	done chan tea.Msg
}

// Tick cria um comando de temporização ligado ao relógio virtual da
// simulação, como alternativa a tea.Tick.
//
// Diferente de tea.Tick, o comando não espera tempo real: a mensagem só é
// produzida quando o relógio virtual avança (veja [Simulator.Advance]). Use
// nos modelos que você quer testar sem consumo de tempo real.
//
// Exemplo:
//
//	func (m model) Init() tea.Cmd {
//	    return charmtest.Tick(time.Second, func(t time.Time) tea.Msg {
//	        return tickMsg(t)
//	    })
//	}
func Tick(d time.Duration, fn func(time.Time) tea.Msg) tea.Cmd {
	t := &virtualTimer{dur: d, fn: fn}
	return func() tea.Msg { return t }
}

// Every cria um comando de temporização ligado ao relógio virtual da
// simulação, alinhado aos múltiplos do intervalo, como tea.Every.
//
// Assim como [Tick], a mensagem só é produzida quando o relógio virtual
// avança.
func Every(d time.Duration, fn func(time.Time) tea.Msg) tea.Cmd {
	t := &virtualTimer{dur: d, every: true, fn: fn}
	return func() tea.Msg { return t }
}

// Ponteiros de código das funções de temporização do Bubble Tea. Todas as
// closures produzidas por tea.Tick compartilham o mesmo ponteiro de código,
// o que permite reconhecê-las por reflection mesmo sem acesso ao tipo
// (que não é exportado).
var (
	tickCodePtr  = reflect.ValueOf(tea.Tick(time.Hour, func(time.Time) tea.Msg { return nil })).Pointer()
	everyCodePtr = reflect.ValueOf(tea.Every(time.Hour, func(time.Time) tea.Msg { return nil })).Pointer()
)

// isBlockingTimer informa se o comando é um temporizador do Bubble Tea
// (tea.Tick ou tea.Every). Esses comandos bloqueiam até o prazo e, por isso,
// precisam ser adiados: executá-los de forma síncrona travaria a simulação
// (e o padrão "Init() devolve tea.Tick" nunca terminaria).
func isBlockingTimer(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	ptr := reflect.ValueOf(cmd).Pointer()
	return ptr == tickCodePtr || ptr == everyCodePtr
}

// sequenceCmds extrai os comandos de uma tea.Sequence.
//
// O Bubble Tea v2 usa um tipo não exportado (sequenceMsg) para representar
// uma sequência de comandos. Como o tipo não é acessível, ele é reconhecido
// por reflection: o nome do tipo, o pacote e o tipo dos elementos precisam
// bater com uma fatia de tea.Cmd.
func sequenceCmds(msg tea.Msg) ([]tea.Cmd, bool) {
	if msg == nil {
		return nil, false
	}
	t := reflect.TypeOf(msg)
	if t.Kind() != reflect.Slice || t.Name() != "sequenceMsg" {
		return nil, false
	}
	if t.Elem() != reflect.TypeOf(tea.Cmd(nil)) {
		return nil, false
	}
	v := reflect.ValueOf(msg)
	cmds := make([]tea.Cmd, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		cmd, ok := v.Index(i).Interface().(tea.Cmd)
		if !ok {
			return nil, false
		}
		cmds = append(cmds, cmd)
	}
	return cmds, true
}

// Now retorna o instante atual do relógio virtual.
func (s *Simulator) Now() time.Time {
	return s.now
}

// Advance avança o relógio virtual em d e entrega as mensagens dos
// temporizadores vencidos.
//
// Temporizadores virtuais (criados por [Tick] e [Every]) disparam
// instantaneamente, sem consumir tempo real. Temporizadores reais do Bubble
// Tea (tea.Tick e tea.Every) dependem do timer do sistema operacional: o
// simulador espera o disparo dentro da janela avançada, limitado por d de
// tempo real.
//
// Exemplo:
//
//	sim := charmtest.New(spinnerModel{})
//	sim.Advance(3 * time.Second) // três ticks do spinner
func (s *Simulator) Advance(d time.Duration) {
	if d < 0 {
		s.setErr(fmt.Errorf("charmtest: Advance não aceita duração negativa (%s)", d))
		return
	}
	s.now = s.now.Add(d)
	s.scheduleBase = s.now
	s.fireVirtualTimers()
	s.fireRealTimers(d)
}

// AdvanceUntilIdle entrega as mensagens dos temporizadores que já estão
// prontos (temporizadores virtuais vencidos e temporizadores reais que já
// dispararam) e processa a cascata de reações, sem avançar o relógio
// virtual.
func (s *Simulator) AdvanceUntilIdle() {
	for i := 0; i < s.maxDepth; i++ {
		fired := s.fireVirtualTimers()
		if s.drainReadyRealTimers() {
			fired = true
		}
		if !fired {
			return
		}
	}
	s.setErr(fmt.Errorf("charmtest: AdvanceUntilIdle excedeu o limite de %d iterações; o modelo não fica ocioso", s.maxDepth))
}

// PendingTimers retorna a quantidade de temporizadores ainda pendentes
// (virtuais e reais).
func (s *Simulator) PendingTimers() int {
	return len(s.timers) + len(s.realTimers)
}

// scheduleVirtualTimer agenda um temporizador virtual no relógio virtual.
func (s *Simulator) scheduleVirtualTimer(t *virtualTimer) {
	if t.dur <= 0 {
		s.deliverVirtualTimer(t)
		return
	}
	// O prazo é calculado a partir da base de agendamento, que normalmente é
	// o instante atual do relógio virtual. Durante a reação a um timer
	// atrasado, a base é o prazo do timer, para que o próximo ciclo do
	// modelo continue alinhado à grade original.
	base := s.scheduleBase
	t.due = base.Add(t.dur)
	if t.every {
		// tea.Every dispara alinhado ao relógio; o virtual segue a mesma regra.
		t.due = base.Truncate(t.dur).Add(t.dur)
		if !t.due.After(base) {
			t.due = t.due.Add(t.dur)
		}
	}
	s.timers = append(s.timers, t)
}

// deliverVirtualTimer produz a mensagem do temporizador e a entrega ao modelo.
func (s *Simulator) deliverVirtualTimer(t *virtualTimer) {
	if t.fn == nil {
		return
	}
	if msg := t.fn(s.scheduleBase); msg != nil {
		s.Send(msg)
	}
}

// fireVirtualTimers dispara todos os temporizadores virtuais vencidos,
// processando as reações em cascata. Informe true se algo foi disparado.
func (s *Simulator) fireVirtualTimers() bool {
	fired := false
	for s.events < s.maxEvents {
		var next *virtualTimer
		for _, t := range s.timers {
			if t.due.After(s.now) {
				continue
			}
			if next == nil || t.due.Before(next.due) {
				next = t
			}
		}
		if next == nil {
			return fired
		}
		s.removeVirtualTimer(next)
		fired = true

		// A reação do modelo acontece no instante do timer, não no instante
		// final do avanço.
		base := s.scheduleBase
		s.scheduleBase = next.due
		s.deliverVirtualTimer(next)
		s.scheduleBase = base
	}
	s.setErr(fmt.Errorf("charmtest: limite de %d eventos excedido durante o avanço do relógio virtual", s.maxEvents))
	return fired
}

// removeVirtualTimer remove um temporizador virtual da fila.
func (s *Simulator) removeVirtualTimer(target *virtualTimer) {
	for i, t := range s.timers {
		if t == target {
			s.timers = append(s.timers[:i], s.timers[i+1:]...)
			return
		}
	}
}

// scheduleRealTimer executa um temporizador real em uma goroutine. O timer do
// Bubble Tea começa a contar assim que o comando é criado (dentro de
// Update), então a execução é iniciada imediatamente, como em produção.
func (s *Simulator) scheduleRealTimer(cmd tea.Cmd) {
	t := &realTimer{done: make(chan tea.Msg, 1)}
	go func() {
		t.done <- cmd()
	}()
	s.realTimers = append(s.realTimers, t)
}

// fireRealTimers entrega as mensagens dos temporizadores reais que disparam
// dentro do orçamento de tempo real informado.
func (s *Simulator) fireRealTimers(budget time.Duration) bool {
	if len(s.realTimers) == 0 {
		return false
	}
	deadline := time.Now().Add(budget)
	fired := false
	for len(s.realTimers) > 0 && s.events < s.maxEvents {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fired
		}
		t := s.realTimers[0]
		select {
		case msg := <-t.done:
			s.realTimers = s.realTimers[1:]
			fired = true
			if msg != nil {
				s.Send(msg)
			}
		case <-time.After(remaining):
			return fired
		}
	}
	return fired
}

// drainReadyRealTimers entrega as mensagens dos temporizadores reais que já
// dispararam, sem esperar.
func (s *Simulator) drainReadyRealTimers() bool {
	fired := false
	for len(s.realTimers) > 0 && s.events < s.maxEvents {
		select {
		case msg := <-s.realTimers[0].done:
			s.realTimers = s.realTimers[1:]
			fired = true
			if msg != nil {
				s.Send(msg)
			}
		default:
			return fired
		}
	}
	return fired
}
