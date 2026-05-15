/*
 * Copyright 2014 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package grpc implements an RPC system called gRPC.
package grpc

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The format of the payload: compressed or not?
type payloadFormat uint8

const (
	compressionNone payloadFormat = 0 // no compression
	compressionMade payloadFormat = 1 // compressed
)

// parser reads complete gRPC messages from the underlying reader.
type parser struct {
	// r is the underlying reader.
	r io.Reader

	// The header of a gRPC message. Find more detail at
	// https://grpc.io/docs/what-is-grpc/core-concepts/
	header [5]byte
}

// recvMsg reads a complete gRPC message from the stream.
// It returns the message payload and any error encountered.
func (p *parser) recvMsg(maxReceiveMessageSize int) (pf payloadFormat, msg []byte, err error) {
	if _, err := io.ReadFull(p.r, p.header[:]); err != nil {
		return 0, nil, err
	}

	pf = payloadFormat(p.header[0])
	length := binary.BigEndian.Uint32(p.header[1:])

	if length == 0 {
		return pf, nil, nil
	}
	if int64(length) > int64(maxInt) {
		return 0, nil, status.Errorf(codes.ResourceExhausted, "grpc: received message larger than max length allowed on current machine (%d vs. %d)", length, maxInt)
	}
	if int(length) > maxReceiveMessageSize {
		return 0, nil, status.Errorf(codes.ResourceExhausted, "grpc: received message larger than max (%d vs. %d)", length, maxReceiveMessageSize)
	}

	msg = make([]byte, int(length))
	if _, err := io.ReadFull(p.r, msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	return pf, msg, nil
}

// The default maximum receive and send message size.
const (
	defaultClientMaxRecvMsgSize = 1024 * 1024 * 4  // 4MB
	defaultClientMaxSendMsgSize = math.MaxInt32
	defaultServerMaxRecvMsgSize = 1024 * 1024 * 4  // 4MB
	defaultServerMaxSendMsgSize = math.MaxInt32
)

// maxInt is the maximum int value on the current platform.
const maxInt = int(^uint(0) >> 1)

// gzip compressor pool to reduce allocations.
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, err := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		if err != nil {
			panic(fmt.Sprintf("grpc: failed to create gzip writer: %v", err))
		}
		return w
	},
}

// encode serializes msg and returns the payload bytes. If msg is nil,
// it returns an empty byte slice.
func encode(msg interface{ Marshal() ([]byte, error) }) ([]byte, error) {
	if msg == nil {
		return nil, nil
	}
	b, err := msg.Marshal()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "grpc: error while marshaling: %v", err.Error())
	}
	if uint(len(b)) > uint(maxInt) {
		return nil, status.Errorf(codes.ResourceExhausted, "grpc: message too large (%d bytes)", len(b))
	}
	return b, nil
}

// msgHeader returns a 5-byte header for the gRPC wire format.
// The first byte indicates compression (0 = none, 1 = compressed).
// The next 4 bytes encode the message length as a big-endian uint32.
func msgHeader(data []byte, pf payloadFormat) []byte {
	h := make([]byte, 5)
	h[0] = byte(pf)
	binary.BigEndian.PutUint32(h[1:], uint32(len(data)))
	return h
}
