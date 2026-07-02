package messaging

import (
	"context"
	"errors"
	"sync"
	"time"
)

// MessageBus define as capacidades do barramento interno de mensagens do servidor.
type MessageBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan any
}

var (
	Instance *MessageBus
	once     sync.Once
)

// GetInstance retorna o singleton thread-safe do MessageBus.
func GetInstance() *MessageBus {
	once.Do(func() {
		Instance = &MessageBus{
			subscribers: make(map[string][]chan any),
		}
	})
	return Instance
}

// Subscribe inscreve-se em um tópico e retorna um canal somente leitura de mensagens.
func (mb *MessageBus) Subscribe(topic string) <-chan any {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ch := make(chan any, 100) // Buffer de 100 mensagens para evitar bloqueios rápidos
	mb.subscribers[topic] = append(mb.subscribers[topic], ch)
	return ch
}

// Publish envia de forma não bloqueante uma mensagem para todos os assinantes ativos do tópico.
func (mb *MessageBus) Publish(topic string, message any) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	subs, exists := mb.subscribers[topic]
	if !exists {
		return
	}

	for _, ch := range subs {
		select {
		case ch <- message:
		default:
			// Canal cheio, descarta ou trata slow consumer para evitar deadlock completo do engine
		}
	}
}

// RequestReply executa um padrão síncrono de pergunta e resposta usando canais temporários.
func (mb *MessageBus) RequestReply(topic string, request any, timeout time.Duration) (any, error) {
	// Cria um ID de resposta único para este request
	replyTopic := topic + ".reply." + string(time.Now().UnixNano())

	replyChan := mb.Subscribe(replyTopic)
	defer mb.Unsubscribe(replyTopic, replyChan)

	// Encapsula o request original informando para onde enviar a resposta
	envelope := struct {
		Payload    any
		ReplyTopic string
	}{
		Payload:    request,
		ReplyTopic: replyTopic,
	}

	mb.Publish(topic, envelope)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case response := <-replyChan:
		return response, nil
	case <-ctx.Done():
		return nil, errors.New("request-reply operation timed out")
	}
}

// Unsubscribe remove um canal de inscrição de um tópico específico.
func (mb *MessageBus) Unsubscribe(topic string, ch <-chan any) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	subs, exists := mb.subscribers[topic]
	if !exists {
		return
	}

	for i, sub := range subs {
		if sub == ch {
			// Fecha o canal de escrita para limpar recursos
			close(sub)
			mb.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	if len(mb.subscribers[topic]) == 0 {
		delete(mb.subscribers, topic)
	}
}
