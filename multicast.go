// Copyright 2026 James Cunningham
// SPDX-License-Identifier: BSD-3-Clause
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file or at https://opensource.org/license/BSD-3-clause

package netlink

import (
	"fmt"
	"os"
	"sync"
)

// Subscription is a Netlink connection configured to receive asynchronous
// multicast notifications from one-or-more groups in a Netlink family.
//
// It is safe for concurrent use.
type Subscription interface {
	// Close the subscription, which closes the underlying socket, and no more
	// messages will be received.
	//
	// Any reads in-flight will return [ErrClosed].
	Close() error

	// Family returns the Netlink [Family] this subscription was opened with.
	Family() Family

	// Groups returns a list of the groups subscribed to.
	Groups() []int

	// ReceiveMessage will block until a message is read from the underlying
	// [Conn], which will be unmarshaled into the given [Unmarshaler].
	//
	// It is safe to call this concurrently to handle many massages.
	//
	// If [Subscription.Close] is called while this is waiting for a message,
	// [ErrClosed] will be returned.
	ReceiveMessage(Unmarshaler) error
}

type subscription struct {
	nl     Conn
	groups []int
	pool   *sync.Pool
}

// Subscribe opens an connection to a Netlink socket with the specified
// [Family], then subscribes to the given group IDs for multicast
// notifications.
//
// Once configured, the [Subscription.ReceiveMessage] can be used to receive
// the messages, until closed.
func Subscribe(family Family, groups ...int) (Subscription, error) {
	if len(groups) == 0 {
		return nil, fmt.Errorf("no groups to subscribe to")
	}

	nl, err := Dial(family, JoinGroups(groups...))
	if err != nil {
		return nil, err
	}

	return &subscription{
		nl:     nl,
		groups: groups,
		pool: &sync.Pool{
			New: func() any {
				b := make([]byte, os.Getpagesize())
				return &b
			},
		},
	}, nil
}

func (s *subscription) Close() error {
	return s.nl.Close()
}

func (s *subscription) Family() Family {
	return s.nl.Family()
}

func (s *subscription) Groups() []int {
	return s.groups
}

func (s *subscription) ReceiveMessage(dst Unmarshaler) error {
	bufp, _ := s.pool.Get().(*[]byte)
	defer s.pool.Put(bufp)

	n, err := s.nl.Read(*bufp)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	msg, err := NewMessageDecoder((*bufp)[:n])
	if err != nil {
		return err
	}

	err = dst.UnmarshalNetlink(msg)
	if err != nil {
		return err
	}

	return nil
}
