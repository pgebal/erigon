// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/rlp"
	"github.com/stretchr/testify/assert"
)

// Tests that the custom union field encoder and decoder works correctly.
func TestGetBlockHeadersDataEncodeDecode(t *testing.T) {
	// Create a "random" hash for testing
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	// Assemble some table driven tests
	tests := []struct {
		packet *GetBlockHeadersPacket
		fail   bool
	}{
		// Providing the origin as either a hash or a number should both work
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Number: 314}}},
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Hash: hash}}},

		// Providing arbitrary query field should also work
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Number: 314}, Amount: 314, Skip: 1, Reverse: true}},
		{fail: false, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Hash: hash}, Amount: 314, Skip: 1, Reverse: true}},

		// Providing both the origin hash and origin number must fail
		{fail: true, packet: &GetBlockHeadersPacket{Origin: HashOrNumber{Hash: hash, Number: 314}}},
	}
	// Iterate over each of the tests and try to encode and then decode
	for i, tt := range tests {
		bytes, err := rlp.EncodeToBytes(tt.packet)
		if err != nil && !tt.fail {
			t.Fatalf("test %d: failed to encode packet: %v", i, err)
		} else if err == nil && tt.fail {
			t.Fatalf("test %d: encode should have failed", i)
		}
		if !tt.fail {
			packet := new(GetBlockHeadersPacket)
			if err := rlp.DecodeBytes(bytes, packet); err != nil {
				t.Fatalf("test %d: failed to decode packet: %v", i, err)
			}
			if packet.Origin.Hash != tt.packet.Origin.Hash || packet.Origin.Number != tt.packet.Origin.Number || packet.Amount != tt.packet.Amount ||
				packet.Skip != tt.packet.Skip || packet.Reverse != tt.packet.Reverse {
				t.Fatalf("test %d: encode decode mismatch: have %+v, want %+v", i, packet, tt.packet)
			}
		}
	}
}

// TestEth66EmptyMessages tests encoding of empty eth66 messages
func TestEth66EmptyMessages(t *testing.T) {
	// All empty messages encodes to the same format
	want := common.FromHex("c4820457c0")

	for i, msg := range []interface{}{
		// Headers
		GetBlockHeadersPacket66{1111, nil},
		BlockHeadersPacket66{1111, nil},
		// Bodies
		GetBlockBodiesPacket66{1111, nil},
		BlockBodiesPacket66{1111, nil},
		BlockBodiesRLPPacket66{1111, nil},
		// Node data
		GetNodeDataPacket66{1111, nil},
		NodeDataPacket66{1111, nil},
		// Receipts
		GetReceiptsPacket66{1111, nil},
		ReceiptsPacket66{1111, nil},
		// Transactions
		GetPooledTransactionsPacket66{1111, nil},
		PooledTransactionsPacket66{1111, nil},
		PooledTransactionsRLPPacket66{1111, nil},

		// Headers
		BlockHeadersPacket66{1111, BlockHeadersPacket([]*types.Header{})},
		// Bodies
		GetBlockBodiesPacket66{1111, GetBlockBodiesPacket([]common.Hash{})},
		BlockBodiesPacket66{1111, BlockBodiesPacket([]*BlockBody{})},
		BlockBodiesRLPPacket66{1111, BlockBodiesRLPPacket([]rlp.RawValue{})},
		// Node data
		GetNodeDataPacket66{1111, GetNodeDataPacket([]common.Hash{})},
		NodeDataPacket66{1111, NodeDataPacket([][]byte{})},
		// Receipts
		GetReceiptsPacket66{1111, GetReceiptsPacket([]common.Hash{})},
		ReceiptsPacket66{1111, ReceiptsPacket([][]*types.Receipt{})},
		// Transactions
		GetPooledTransactionsPacket66{1111, GetPooledTransactionsPacket([]common.Hash{})},
		PooledTransactionsPacket66{1111, PooledTransactionsPacket([]types.Transaction{})},
		PooledTransactionsRLPPacket66{1111, PooledTransactionsRLPPacket([]rlp.RawValue{})},
	} {
		if have, _ := rlp.EncodeToBytes(msg); !bytes.Equal(have, want) {
			t.Errorf("test %d, type %T, have\n\t%x\nwant\n\t%x", i, msg, have, want)
		}
	}

}

// TestEth66Messages tests the encoding of all redefined eth66 messages
func TestEth66Messages(t *testing.T) {

	// Some basic structs used during testing
	var (
		header       *types.Header
		blockBody    *BlockBody
		blockBodyRlp rlp.RawValue
		txs          []types.Transaction
		txRlps       []rlp.RawValue
		hashes       []common.Hash
		receipts     []*types.Receipt
		receiptsRlp  rlp.RawValue

		err error
	)
	header = &types.Header{
		Difficulty: big.NewInt(2222),
		Number:     big.NewInt(3333),
		GasLimit:   4444,
		GasUsed:    5555,
		Time:       6666,
		Extra:      []byte{0x77, 0x88},
	}
	// Init the transactions, taken from a different test
	{
		for _, hexrlp := range []string{
			"f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10",
			"f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afb",
		} {
			var tx types.Transaction
			rlpdata := common.FromHex(hexrlp)
			tx, err1 := types.DecodeTransaction(rlpdata)
			if err1 != nil {
				t.Fatal(err1)
			}
			txs = append(txs, tx)
			txRlps = append(txRlps, rlpdata)
		}
	}
	// init the block body data, both object and rlp form
	blockBody = &BlockBody{
		Transactions: txs,
		Uncles:       []*types.Header{header},
	}
	blockBodyRlp, err = rlp.EncodeToBytes(blockBody)
	if err != nil {
		t.Fatal(err)
	}

	hashes = []common.Hash{
		common.HexToHash("deadc0de"),
		common.HexToHash("feedbeef"),
	}
	byteSlices := [][]byte{
		common.FromHex("deadc0de"),
		common.FromHex("feedbeef"),
	}
	// init the receipts
	{
		receipts = []*types.Receipt{
			{
				Status:            types.ReceiptStatusFailed,
				CumulativeGasUsed: 1,
				Logs: []*types.Log{
					{
						Address: common.BytesToAddress([]byte{0x11}),
						Topics:  []common.Hash{common.HexToHash("dead"), common.HexToHash("beef")},
						Data:    []byte{0x01, 0x00, 0xff},
					},
				},
				TxHash:          hashes[0],
				ContractAddress: common.BytesToAddress([]byte{0x01, 0x11, 0x11}),
				GasUsed:         111111,
			},
		}
		rlpData, err := rlp.EncodeToBytes(receipts)
		if err != nil {
			t.Fatal(err)
		}
		receiptsRlp = rlpData
	}

	for i, tc := range []struct {
		message interface{}
		want    []byte
	}{
		{
			GetBlockHeadersPacket66{1111, &GetBlockHeadersPacket{HashOrNumber{hashes[0], 0}, 5, 5, false}},
			common.FromHex("e8820457e4a000000000000000000000000000000000000000000000000000000000deadc0de050580"),
		},
		{
			GetBlockHeadersPacket66{1111, &GetBlockHeadersPacket{HashOrNumber{common.Hash{}, 9999}, 5, 5, false}},
			common.FromHex("ca820457c682270f050580"),
		},
		{
			BlockHeadersPacket66{1111, BlockHeadersPacket{header}},
			common.FromHex("f90202820457f901fcf901f9a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008208ae820d0582115c8215b3821a0a827788a00000000000000000000000000000000000000000000000000000000000000000880000000000000000"),
		},
		{
			GetBlockBodiesPacket66{1111, GetBlockBodiesPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			BlockBodiesPacket66{1111, BlockBodiesPacket([]*BlockBody{blockBody})},
			common.FromHex("f902dc820457f902d6f902d3f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afbf901fcf901f9a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008208ae820d0582115c8215b3821a0a827788a00000000000000000000000000000000000000000000000000000000000000000880000000000000000"),
		},
		{ // Identical to non-rlp-shortcut version
			BlockBodiesRLPPacket66{1111, BlockBodiesRLPPacket([]rlp.RawValue{blockBodyRlp})},
			common.FromHex("f902dc820457f902d6f902d3f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afbf901fcf901f9a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000940000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008208ae820d0582115c8215b3821a0a827788a00000000000000000000000000000000000000000000000000000000000000000880000000000000000"),
		},
		{
			GetNodeDataPacket66{1111, GetNodeDataPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			NodeDataPacket66{1111, NodeDataPacket(byteSlices)},
			common.FromHex("ce820457ca84deadc0de84feedbeef"),
		},
		{
			GetReceiptsPacket66{1111, GetReceiptsPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			ReceiptsPacket66{1111, ReceiptsPacket([][]*types.Receipt{receipts})},
			common.FromHex("f90172820457f9016cf90169f901668001b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f85ff85d940000000000000000000000000000000000000011f842a0000000000000000000000000000000000000000000000000000000000000deada0000000000000000000000000000000000000000000000000000000000000beef830100ff"),
		},
		{
			ReceiptsRLPPacket66{1111, ReceiptsRLPPacket([]rlp.RawValue{receiptsRlp})},
			common.FromHex("f90172820457f9016cf90169f901668001b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f85ff85d940000000000000000000000000000000000000011f842a0000000000000000000000000000000000000000000000000000000000000deada0000000000000000000000000000000000000000000000000000000000000beef830100ff"),
		},
		{
			GetPooledTransactionsPacket66{1111, GetPooledTransactionsPacket(hashes)},
			common.FromHex("f847820457f842a000000000000000000000000000000000000000000000000000000000deadc0dea000000000000000000000000000000000000000000000000000000000feedbeef"),
		},
		{
			PooledTransactionsPacket66{1111, PooledTransactionsPacket(txs)},
			common.FromHex("f8d7820457f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afb"),
		},
		{
			PooledTransactionsRLPPacket66{1111, PooledTransactionsRLPPacket(txRlps)},
			common.FromHex("f8d7820457f8d2f867088504a817c8088302e2489435353535353535353535353535353535353535358202008025a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c12a064b1702d9298fee62dfeccc57d322a463ad55ca201256d01f62b45b2e1c21c10f867098504a817c809830334509435353535353535353535353535353535353535358202d98025a052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afba052f8f61201b2b11a78d6e866abc9c3db2ae8631fa656bfe5cb53668255367afb"),
		},
	} {
		if have, _ := rlp.EncodeToBytes(tc.message); !bytes.Equal(have, tc.want) {
			t.Errorf("test %d, type %T, have\n\t%x\nwant\n\t%x", i, tc.message, have, tc.want)
		}
	}
}
func TestDecodePooledTx(t *testing.T) {
	in := []string{
		"f90420a0e2c35c05dbb07cf2e0784116966bd5d582846bdb54c928e740492262ccfeb40ca0c4fd5ae5baa1180ff6dd8c4b01cc2d93f8907de27ac2016648c6373478bef503a0b35dca76c30d732374041438c26b9ee8f643744446ab53c7f38fa2074eaac992a08518fdec2a229d82884b323937f4a0f09b02ac746aaf6773d482e40b597c05c7a08f8d9110ea6acca9554e2f80cac3cedbb143fbe4a1ab3e60ed31f4c74d748deca016b0027c133e3372deffb4e86e1bc5e975d1d610a9f2fdea3a74e4d60ef864baa00c6332c83e85ebb9a3a7afd313df4a82c7e5decb5872e264a3f11da610245f00a08bedc14cf929b536e30df8538b870518016096a4870c7c20e5b48da7b03df384a05a9ff87469a5332cbd18520dcdf25c17425de71d9efb56072e023246617145ffa08a165cd6bdc433f6b97e8a6e204ad59cbee4084421aa810a1e2cf9c828e949e8a09521157629d905e3337a9b54522709ccf8890cfd717def560b053926c1862ed3a0284d676704a205a5c932dba13672caafe13a727b3a3cd7dc56bc49a6c6c9a7e8a09f12e7753596116fd9ebe554428f927e8afc99bad30ad28cd9deeeaebe21cd99a09497716972c5f7ff311769f185f64adac8c54f096cab5beb7ad80cfc7e773ccfa00e0036fd7086e15f045b11d7a22df7dc99958124ccc7100d499d182becb7ed59a000d8094ecdee29e6c916360f1594d68cda566d9bd2a17a19312eb0487d46b537a03b1bc9e360c01b2d6cf9ba42a06b91b30692a086daf3ba6db2b661cfd5f1fb61a08cdd05c1605609f1979fd55ff7de0a16e41decb7ab3e340b9b729a88ea8cbb5ca0151bc4c1e0cbd896cabdaf5d08830cb608337ac247347786b2191f79bd367d96a0c9d895b6094195745740ff05e2fd4e55b33ae554843cabc70ad64b71a1a8093ca0ffdf8f28142597911de68d59e7f73209aa2573583bc3c2aaa144e2727153a1aea0535a8ccc5eb9f69005021093017da51b54289e24fb24f8f36ee8d421e3df99b7a0b0428184b221db9f5bd7c2c3ab1b4e21cf961f9218bce2308a1f1be2da40195ba0294f7127f5306e5a5a57127b1300351dac1affef78fb7e822fab1175e021c261a03b2d6b186168a6ec0a364ff65e58b34c1b4d85cf0d30cf3158c68b3ff9987621a0eba6b975d9d964dd2db4f0e562cb16ca1b51d42c659df15f28a84c31e2fb7a9aa00d9b8879c571278d965e847752f9b97b12a965b44bfb1a653fa0b983949bad8da0752d870c51884688996ebe5c561f7b05602fe7939b5f3651fb2a848921d3a88ea0bc73d7980607d967aca0c8b07a44a62fcc6f5813587987cf65d972396e45b448a0c3fc9e18bf84f8dbc72f1e5d106f760055982ceabb0fe9b62c37cb73ee0d04d1a083bcb5e1bf1871154a16c9d50c40656feeca78ba66d5f2851a031952f2e4f96ba085fe3f02e443d6be1ddbe9ea3949045e9787350caa744f9432e92046d2331e86",
		"f90252a08300b9b4eafc34783aa73c7cb838a57dfc67297a3b1d8257246bd185a4f93b3aa02f34818121bf0ba43a72b792c082958f3d5f228b21a2eac269817fb3599a84e2a063aa50744b161d468fe18ea34dae296dad9cb12dd379817705381d891fdc5597a0ad9655f53ba172b838cf806abe28c77f3b264403b67ae1542a45fe37b7d4a863a03509363a668176aa2199689909c506996aaef511a41259d8f76768c4bb6d4752a0d1701f189b824587e22751ea6d58a7c5c1f7c80c5b09e3f3f92661ff32acb8f6a014f08af42770974caca620efa32c794d8a36496c873bf1921c12be196d096579a0c7377d2c6550bb5dc4bbffba41cac186c33b2d6252d1e0e24f5b0a0bace59472a02c2ba0c1f3555dc2256e528751f30ccc52112359258a70bb0a8fc18ece48381ca07d614cb5a2b817db6513993c8e7bbd6c1797e8e9bda6682f5e1df513ed2034f2a03b4b85d126965773ceb1786b11f2c5a949a87b0f6fe8d872ad89cccfd1f87ed1a0656647ae904635ff87dbaa5ca911a3be660932170eba4a21d0dbdbe97d20eeeaa0389b40fccf0dbb545062f2191469605c1840adf1f0a0a894813220fcacf64159a00e9fee2d2799359f2f30fdf4730b975a8c2bd88580bcae9ada7207727307dad9a03ba2575ccf8cd5e5cfb3a49ee7e7eca5d848769633eef17555568bcd6f67dfcaa047257e61f349efa356983cafa152f5709b7e98d237aa1e56909f9e68e3c570dda046afb04e1863bee5f4b2295df87345cf2c57634b5009c6c4422a6f0b2585f9fea015f91ab0306be4a339b8d273e574657112543d3414bbcb1541ed087e950c5732",
		"f92100a0457dbea9ad7a3ac599afe2918a92588ce4f09a7f0fcad832386e455143d3efc1a0e68ce70766ffa9b0dcb672d5806289c6ac94fbbc86702bd99778bd344c7311c9a0416da3c5b9d7ff5534102eefd90fac19e3501f730be4875496c79690f1250887a0db85964d4463f5c06c86f038b5ef19ee16be469b120e875ba0d1a7e9d2c9914da03a77f75d108289f8318601d27a18e149efee678f5438e48ee0b3242242de9861a04084aca085236dde3b4c21bf2167a44a550d2378d1b157472a3d114c37bf3f74a0f7e3253460b1065766a297e2df2e9ac33f844b3b3512b5b7a8df427d41ce27e2a050f68e46ed07feee900ffeaa670fee83bd4d47bb3deba88178de2ae09b7c8216a0e0a3cdbf78229fa645f7f7b6bf77bb68f4bfc28f2064d8d059f591af1ba0e1f5a05f64e70da5f5c60d38c3cecd908e1adcaf546fd430558fcfc023680e8b4f29c7a0de12e201d4c361d34f5bf6f6d4beebb32fc2494499155a1cb765c9e94a3fb4efa0ca79e282d9b5f78160c2b4bd3ca60554c9d24b778f44cd5462dbbdf2d54b5f01a013d6cddcc76c44e139e815376d7abe6ab641b36f3157e6649cabfec2880414c9a0b886f616f5ff7ee6500a9338968f074df49aa11971660637e3e34492358ee47ea0ee99c6fbdf5662030fef6425efcbfa8078f0f431fde3be1b8aae4aea3ba580bda07f1a853063aa035138b76ed0135110b5a10ea8d0e47ed31bf2a591042a13707da0a72b949c7db6ec8a676f86c7e3d8fa598dce0ce5949db92e5b82cb0151633098a000df9aaf75db347c35895b51f9e066b3fa49950b4a8ef81ed4bf8acbef7f3075a07945ece5f3c656551faf4d1ded4e3c7b392a3b85c662b30544a6a3f04a57643fa0055b598c472aebcf8028244900926e2d94b9caa7f8e67b2a0d88e0cffa9e7cfba0a8083261ec9d81591b2b1837f4bd756e509637191ada6e08fe6b3cd42de08de4a0d0b17671335c7173cc7aafe9fa9af8b7e52ec056d459547ae627957c332b2d0fa08f8fa96861da8464ee1aebc7c106eeaf3630e69d3f3490de77ef415f2472700ea06b370755efc2aa04857953e4697f487cd26c0f85d7b1c1a223b26f0cd186a9d0a00484268c6711b9455ba4b48e1b6f5e88b1dad5e77971ead0f5e68a1c7748c367a07b0a2946456ce64edb7c7361d518ec90b0630884d14bebee345db053781dfa17a01aaf2e5e97053a511044aa31477fbbc22d1f6dee5c0ccc830c4a6240fc4d6484a0540f6302b6900a3403af090b0439a9cc9441d6518a58001ded05693dfddcc7b0a0cd989d5ce499b59cadcbf8c725cc2faf5aa21d66915d505914d7ad0aa6f8de2ea04e49fc458aaf05fe5c017d7c48bd02a246921086ab509dfd66546b1ee51b434fa0df2b1cfe7b2c6ab27c058a55432d60c79f02abe1c9c26e360e5c9819a89e7918a0cf076e52950132ae35163f836c378bded5303ea7e857ece11baf28ac5b1c98aea0f9d36cc0621b69d71e85f48a949f7c388cc3a920760c76705fc70b46c7ec338ea0a72ac1227bacd06bd244249d2b7c58ca96748d6028860849d07cc8a1591d7cb6a0a2d3bc56974ab12c9b26e649b333f6a0e9d6c92a7e9d25eee725f915892f7139a099b67b2e993f1bc2aed218b9a4919e29688ef06a93d0fd1867f065da154eb161a04f74601c3749db0ce93d86b4721b2dacf5a658367aee6cd96a2b748ac5beee83a0891342c798f6f0fd07ac5c39cd9467ebc2c314a2cd2130fb7622f8d68046caaca08825e19c1ff6fd9c67a9ce3c78ddc80dd5b787a70929562701cdbf785b540f0da01e3d5f79156f0f32620a9d9b3fbe4b80e8124786fd73995bad39a8e0c135ae12a0728b79e3b6d77a24051ac98ae5ee961ea7287c4dc727c428ed637bd87a9de051a0e1e5d2f1c040557a81c93ab0d598be7a81badf8d6baeb378dc9a707764fa2334a0525d7ab8fa251a7dae7900ca104e0a7539eb3cdec22a65f2f947b71e5f6c956fa031720eb289af647dabff6f862a28d2e9b8d5f4102164a86e1858fe0efd174bcda0742aa0021a0b75ee7a93997d275f51a3da3b3f74ab6eb16a95ae6e12e23f8416a05233dae563d95d477c30a30ec026eab0ea53634db75593db8624799b3afe626aa0f5d46e0b4ede7515a12b8a01d7673e115616d98760478f5a41bd7820ab484c26a0297f1c66544bf893349f2763fc49fcdd8e4128de69dd3d3dad7dcc2682df0383a0422d0f4fa436234e72d1759ff9d8e6ea40b8cec7fa6ffb5f3980c9ba024afda1a07258445f1348421cf54b444bb7ccf9c98310782fb06b6e8a4e676f500c1ef934a00830a8ec772df28f54eafa2877515fdf7deca28c628214686a02c9dbfdd84e0aa0d99365c88ca5a5a632bbd4bd9bd5669bc741ec11c06c1760613b47b8ac2bff33a0a54c236eb6ea525b27f911f243d22352b007df8b8c06ab887c5873700a91942da08af147c4e085d12cde5878ae562fa91a299bb6a690b281632d58b06fae9663bea068ff2a28ccd20dbaa322f52e316d77cb3656bd7a0f15bd8eac03c318802e668ba09f1e547d55b063db182ab3b91e7e70d9af30610b4c5004d77a5c38a6615cab6da043681701d663cbb8a5ef553fa7e77b25a1deecb7ace737761309dd0d7da490baa092c353ee88ca09089c897550e68300dab326dbcb0f885feee8f443a932b07292a09acfd8cc59962b4a8e80df5c6721104e9612e7b441e26017d4b3a5e5b2fc213aa02745d5cabffecf871a73680f831458d7e5af0ed35348b6d68fab7e820d5ba93da06730f4166cb883480201397b47d431d16206a960a9a5e9735ed9848d3649fd90a010426bef092103c93a7fa8bf13bec68031788f77b7d8140d478d35b33f867437a0ef08ce558e2bb64d694a5e64b8313d6fc66c1b3ffd48bb031cbae5694580dd1fa06ac7933a0e1e01c0b2f0e9f6be3fbad7ee8183bcbd81532b682e71dfd14c7efba06d78df5495c0465163b88637dd9913cba30eec7b7d2466fdccdaf3475220d960a08868dda844d63222530d8f3d8ad0964af9826b19d259ca3043ab37b5c29c898aa056c41fc0eb3da9b31c784d3ba48f01b469f33df93025901f42b3cf2a45801a93a0dab91a6346667d23c4f2dc2ba5fefbf9c69064458bf4d745b9e5df458bc47729a0bf91c7803f7a6f0c40800ad80a355072fad51fb17f036b1c6018b4eac122c366a03ec19d9c9d75ece9dad232037b2ade4e5919f6e7c9a1dfa7d5be10ac6ece13e1a0d6b29ed45e2cc6eee398b64148ff39fbd4399b633f05c08588457c2daa65a139a0aa969bb7524b1dd2705abdb4ccf91ed7302eb6c0cf1af073860ae86ca5c3e06aa004fa852789a7bfec993fc9564e177f80db5dd00a4748c6342b3679a2d0bffc36a01e3dcf24966b08b355bf126a882900b41509e47523bbc77b2e191ca518a0fe15a06fab12009a9d490cc738b14f3b964b020acfa7525197796dd0222a4ee4a203f4a0966b9fc8caec86efaba02e09fe9bf17129c423e40f345156677bc51043e54798a0c5b485ca74898f3e9c025d41f430f8d2ba7d86e59ed248c391671f24c8894434a064bccacb8a84a451c9c1faaacd55896279581ea3f87ba2aba159611d0693a54aa0413d94148abee1deaab41d3db3541c15f09948118415bb586dd2776ceab8c71ba0ba3948b8e4bfe6a9943472296d13bdec9da21e91ad8ed33aca0572d9f60412aea059bb512969957b22ba85dec033fde835ce34b8f1b084684e574684ea413977eca01c8353e15e462fee048ca1c0bc3d7dbb430ee795ca395e582075fcb868113140a07a29ce5466da30a7f13d6d49f72c87c925feb699cc27fca7d2c8791e707ac926a0c2305e480ed70e9a13b0abba33148ad254cfd41a2db1c6a3f8df8a5fa8ab5417a0f60a1b008093a93b29eba15399d6dc5fb813ef07cf4d0c6a1974f0469d15ca50a013ad4bebb81c5b104cd7f39a75e2c0a0b00d57a20d3a0de890b1560bc8616141a0dce2cce7d0ed35b4bbf672296127217cd1385eee4618cc14c99ab293dbc9c3bda091dfee49f4865b4a5eb365ac4bec7d0af28859c3b1f73826bff80835027afbcca01186acb6582e6514c3278a04d5aa923934acb3a28781811e654f313189b326a7a0191a8daccb6f08ef5b8f1cdcaaf7f7d16150c7e616d64fa6a733e30e03c92ec7a0d886ff68ead70ef7857f2578a4d09509072dbc474a2ee49adf58b2a1ddc12101a04d6ed9dae79ab736ef6f377a6d59a4d23dbf6326832eb23d198ad11edde20d81a0fdbdb62324d8fc480f9dd10dbd330fe8a64f7370cc75b473d2ebe178b5ca8319a05ab97c3f8780d72ad46beb1af9472a836a74c9262e3221586cfc90c1839b93baa04bb7692bfc0bd42f0aaecb2faa7fc9322d267b9109c73c83e8651817db82e229a0106e0f1277527b5161d59186a6ed8a3c720d431e863a6308842e3e7c35cdfa95a04e76483b655fa39ddca148b56e2245604db9c732a27f04b7542c9f22c0e3fe98a0fef5f8c2e54e9bbd82b33debb3a985be5733ff5117325f46a703a03423347d75a0901312021016391e31b0b2b669556700795c86799b6e2536521113d74902ac15a0ac33dfff755cee38b0c1f2ed08a87d5d7ee4facdcb631c95e154332a59113e9aa0f689b57325189a547b7051e34e4921466ef2819225026ea6726dfbb38d481ea5a0ca8ed97e53739e3e0cf71d798e94102c2cc66613785885377b919ed0e5918177a06c0682783f6c7851dd0cf20e2807e5d88ce92c2f439e28f9aedf1cee8c2eb55ea07c726e60fba908cf1846af2c36e8c3be732f954129ebe6f38103518ec22ffa85a03012453f04460ccf8ef43be0b2918200e919a030755fe740c00fc0ef447d26a0a005d6e87101ae85a321d03963b4ed7bd9deb2146ac1275d7e6d7662804d456c8fa08d60aac1234b9daf6b8f6606eda5e744066446e95c83fb4f7b316215191d8f53a0b15da00fe1f119b98c328c3bc0932879bebdc5e59d6df9901dac97b383669acaa0f61cb61a050170400e8c176e445e958cfdf6286b82f3156f9a222db38482fe65a045484737f57987bbd2e1ce6e7fc5e6e82b301dd97bd7fc6422d9dd67d1a6d2f3a053d8371cd24ef9db5f238eec50dcf6dab4284ffdd4882a3e1863f196ff4870bda0dd39aa33f267560be1d88c4bd6a1bf2810338097395d52438dd8e4b02cb84978a079586605fd6153de3d1f1c140c528109e0538f5dc5c63c0d04e4f324589df564a0b5834225c7ed87e0e112f89cd79adeb9d44e314d93cb7b4e21ef277bc3acaa43a0de452afb52a5fd9a20d651ee00bbc348d9bfd23fa69df029b98e3cac50a658fea0662f43b6b2dfc870822fb1c7865f54db04b50679089f9f084647477e86c18917a0ed21bac1b980827276827fee6248b2cea6f8c00a71711e1a95af10e395b11112a08303e88de0a4630e30a95054452cde82856b2ff0ede847719acfb5c06bab5a06a0f06a4b2b27de217b48dd082f0154a9d62f528535ba9f0ffc3a4bedbe00a53cfba0155d26a32b2a3e7b171962da6a30f39335e3954caecc16cb99ef7bce209ff2c9a09691c7d7efb6fe112620fe6ef88d7b6fba53980838cb190da738e0912de3af4aa0eb954bdb5d41c66be4bb38101a9e2ba7e8857d8130b2840a5755c0cc21bd8e0ba0d9667f0cfcc923b80348f189e6f6ca81284e9a1ddda64aef5f346eb5846d3dd5a07de91fa272b5233c007842dd9df5bce1b51776ad9d2bb8b2780023f34e039e63a06680b64e290f34bb395542b117fe27e4f66225f2545d11db1bbd0bd92ca9fca1a02e7c285811f50f35185b3a85d5129a17614de21496b7f5aae5dc5eb55207a167a0b3f63e45c4d670a5bf764d41c6430d6c8a99f3c7191a698f33411a94fc4ac638a04551cb007c5e17f3e9514abe5c4795fb435e63a7e78e71e45dbe3c45fc0f9bd2a0175fd2a244cf6e6ff1247650ee1cd951c81e1d842da9429399a795be4b7366d2a05c39878a681f05c3ff2da0c199afbd37659188df97c983b47ec1a1d7efbb598fa022205aff97dceb64b8a9869fd93325098dd1f2947ab9e8deaefe8dd27b2885f1a00072318dbfcb25b92dd6cb8e5e46824cdb4c9c80019b7de74f7c6e7b0ef4e79ca03eb33c156a4f352f44bbb5d180df4f37c757dcb55ef6bc3406cbd06ce93693b8a0e708d9c6742117260a51d5a4a35013711949d73b16a53437482424cf9cb1d285a0d5b034e2c7782eb827d11119246c85e4be9df88a6cd8581ef15a9c31007b7faca0a23bfb8987002339b017beb01ef612fccc0e2be882d3d13b62789d7ee997a0cfa0ec21e67193e8022dceb7fbc4f519c02e04c12c392d73815a9ab76c18147a8906a0a10cf2794edc9a214ec8af0d7a2d19fcb7aae7b922be1354e31012ab12fea91da00e557cf37dabead6ca434ab7b1e14ceb4e00beecebf84f134bc6241dcfea1065a0683eb17abee39938bbb9fdc8a59b84be1d52a6787345ba0f0fc7290f1e125da9a0998a415478c4d3d852b8d6dfe82b6b172ae40e0789f6c6fb20aca859ce62760ba0f3ca748eef38d808616bfb37ae4475bb8c04a4e4783d7cd1a2edf57dfd9f7306a01f7989f763430c1c1ec3b9b775ce9f1cf2014521b5a01d564cbf678a42d051eaa0e94f98c9099a6cfc66ebb116c233b31b70b1b08c753abae139071262f1816116a0e2da38061116fe543adcba3df34c6245f7310c54e994667fe9930c02b1939dbba0a7d65101a127c971c961a7f6705a48870f7215af5930167aa963aa69ff09846ba0ea35e6495b610495e37ab6fbadc6c04992a602a1797556ee52c59e823189c3e8a0ed699bca16703f0d71f09d72b91bda433f8f3a73ee9440ef6972bac7137c98a8a0b07e3af5035e64940a7ae59001cf9dca43a2f570526b69f7325e610a2280303ea07bb35ca599293bea45afccbdc86a3dcbdbcaac814ec1e3a32a32295ecca5b31ca02252d170dd12c70c5ae3abf9c712d16d28f06276e1df770980392da0488e57d8a073881b9be9e9330aa527c1a92f4d519c42c2252634b1cad345d1c39a4c926641a053b13afff7cb2065f7fd7f0629c74bbf44c38454f51bd59c3153d65e6184ea83a05fbd7fcaf9b67925fa5c07bd121fa6f246eedd2ac1cc9156327520a79bdadee2a0c463853e2b359520f75ea1fcb026a1bbecdc6529c1dcab780b25415caa5e133ea05bc3aeef63a0131e375fa9227dd8203d1e0a1d88f89246cca730c2a093b89bb9a093f9065d1c292cfe90187d8316b5d77f63baa8a3cbacb7dab251e52bb2647992a072caa42fa06337e7a40855ba72ab91bc543af6277b55f45752b2df464aa5980ca0aba9dca288fe67c1a4725c3037b949d1bd601ec66fb08b856a0d0bb3d9746bc6a0c8baeed7aad9f3405c3ec9975af506d0716d9d9c81b0f7583cfb0b4ab23c8411a04549aa50e29845b24bcaccdd73c05ea3f9339fc1b36e674d2d26d1cbebd21707a09ad8a2d70f8fe1bd53e2fe41fb3500cdfe96f2a107391a99e87caff10552bc5aa08274abdd8d2f49a67d2b76a05fc98fe1a6fc61deb8b331d8b9bffe6e5f5b418aa0c44067f13f198366426b116f33f4c4986699834d72235642002687050fc7598fa0079219259c4b66570878e61e68f1ba9bd735db772bd37f1f3a070d3080de6dc7a0dd977bb9b2a8b79015a099ca4fbe1cc03e2bad3a2a5fa6a51556e0a234315b20a00dc885f251fe1d9bb83ea7290bdb2afa4eb7859caa8a5a3c1f09acffc046f931a0a72ebe0ce41209d9d11d05f28e624048089600696a39856e9b99fad176eb1944a05b6118efcf9c6df825e2c11ce1253efa092d9ff62f7973fb9174c982a8261adba06810b59888a32a2b86d66efa776f030cc4d7b47dffbcea30cd59edd04df6c3c2a0f92dbbc2411f5cbe13c11bb174f6d52a15582ab867ad047e17c8b8048ef46073a0a6b6ee59bb9c1ca8bf29789c78bb86d95a426cf7e662149babf37ea1ddd94494a0d41d52b7cf14932699f021dc7d12f29629cfd14bf97ffc448e889f111c59eb25a09cd662bf24415538f1e75e99efed323b3a167488e18d1f06b7518c2cfa8f9f5aa075b09e9c4ddc1675cf3500ebcbddf941d256aeeb746fea4180a5d0b972b38368a0b6c7690d32446df1b5ecda799a68b9e2720999af705b96a59c60e8d5a89ea719a0b59d284f9a25e0c869ac8ab36fc5908b7b511d5bf5dfbe8d4822661bfb7d6215a0caf75cf6533d419e703780fc113a9e8f0b7809e0caa575895ec8953b169af610a00e4beb0e2a7fd498343cdb5d235ee5359405bd759ce4829e5f8ad3ec59250d28a00f0cd100400e2d0f353b9990a73669744951a241bbeda58b81c0d62f2192f405a050ed76f92ef11fd357616ba32cc1174dbad0a859b0a138db4edce0766413321ea033a10af3361aa7900353dbc7439c5623ce3ae18c22e78ae20a885bdc35c7f843a0b91f5bdfc38418e2e675065370d02a7a36adc814c35966849c519c216bdb8af1a01ea59f31ee94703624208b3f87422eebc2852940589bc47d0e5c92afa7c48effa09b20ed326c942ac172c747e28b254f41ba45622a267be543bc9a9eda62dc9a2ca0dda7062f933f07806e2cd3353860f26108e0c1dd4b5975e043715470122818f2a03c405ba0d9b38a1387bf80678d2e1c7e50da7e0440449ff31227ca8debad3728a064ab8d1c3a0a2eb45c9ada53781f44711cab49f8796a6bb2d594379ae5d15027a00fc8ca547ed7b1653eecd6a53f0c59769bc72c6441e1352d58b3c76e07adb105a0afef30ffd851bb7e6a2b4676ac08d62687d5e03fa5890efab7c9922dbd5ecb61a00165d066cabf8c19979ef7253a69680ebb04668d6c5b4897d00792d93351c7dfa01f4c644d845e50fd54608e60d685f4c5e749287d4bbbe43452853268a5b47a0da0d91d97a003adeb26b666f6616092b6dd72b656b1acaabcf91d78bd931fb01884a0c8737e15221df6d9fd90087802abdce3fed99c932cac2593cadb17b9229f9f16a0bed33f5452c61b27d6f1c3f14eecd0092df8526ea5bd8a93461ba27d7f9a75efa0c2b3c6c4869664c2420b0b9f0bbd9718eb6532463949b39f6d345be91c5f9d6fa054d2e5a9275d7c7d8328110bf85631ba36f3a7a21927f840e6a889117171cd9ea044242e49d4c21146e7607350c884a0d3e26e5c3097e1758b482314fe0577aec2a00cd8c8c1ed792f9f6dbfc89b09786c7730b2957d17f79611587aed3fda70d2b8a083fe46d767b5d7e394769fd45f41ef882dcbac6ffb38e96c37eaf6597a42d21aa099e4f39264efb9505d3052fc3eafca2907e2fef0c5b85b0ccc6bc88b6a356c4ea0e4b77f82e45643442713b5c8b46e03f6fb76422177389833290d6bf42a26ab6da03074c90c37ee2a2b219ea7d78d98e6c079ef7bb07aeab5147a302f3366155fb8a05f9ed7bca1374783e7bad9c82a0c353c479ca7e09895d51fed2f5629bcc6abf5a01581b664d619d7627d55862022e4036a76eb2f9cf8a882a40a62f216b990afaca00c3d6f4af686377fd7fcb8c40ac5bc30495f04fe164ee4ee852b9d48425983d7a0c4070b1bd6de7be5e1ea3cb95382f983686b23f4da6f6fd95aa761df6bb90730a02f4725e56a14701a8e1dfef58dad20dca92b4a61a7317789157a388aaf7e6ae5a0f678b9006f0087e46777d82a5b25e11e59ee309925feb9a1d62ac4393082ffcba07460dc419cb17232afb654322ef234f73f8074b819e4c4a6fbdbfc8492392a3da0a5b7291d1afd426b40a99736507c2694865c192f4b436d3d4301fb707a9d5d11a042b9801c4f2615d58ffddf13f716ce02a0650b4f4e40e38d4c6026c0f94ea77fa0fd50fda4772dfee606321698eda27aecd44d3e7bfa130b8e0d6864e305b67ec4a03144dae304be1734ca992969d08316cd0789b55584520c1adca76cf7cab8bceba0d69a4b3e8e8e95f3b8b71aa5b41bd0449c2dad2d46a731a8d347da1127d423cea0babaf47cb6c5d4689316a40c4dd8b382c38191e9a176dc71cea669e9f9cde68ba003cbc12b1c051c2c2c7adfb55cf89e5e32e00f32693736313b1a84bd092b535aa004cce849d28a52d72ba2e4823c51ae543dfed021265c55cb347fbd6f423d8b3ba0126605dee90587eb8e65649f5baa87b826f793df6ff1bcfd91a571c7e81b9838a066fab64e8d0625ede7728e12c42cb56f549caca5b5df21b9fe616b717cda384aa0db8c15c88061ea97e7ae2975c6283011ab2e40396ec8270fba673a2d93da33b8a03b21aae10523f4f023af58b48613d69d103217a9f5a325bb26718c3549d49185a0f5db6576ade94e11674cca382098499d195d70abcd947c63bbcf4efc8ff1b5c8a081995a3a0fbe5b4a60d2f91389b9d49c43501f9107f112c59230d766f1a6e714a09f6bc642a71dfba44745420fd4ffa7af1f8bebbd9c63464bf526757595915f2ea006363434244dde18d1a84dc7e74261534ef4384bd2957a94ad1196788ae1a6a8a09bb434fc8b655cdc080bc47f01de0e0f9bf37c728532e3453dd374af3edc042ea067aa2fd972ef541d3d2401b9be7118bd90e22b3b46682657fc437d6529036fe0a09bba08016ba641e78b95735ead1fa8f470b69fe316a5e2e355eefdb59f0ff6a1a0fc1b26d7ed19e8e4723f1450ad019da802297dd54c446934bb1c1ff971458540a0906ca50e73f23f4647987e0194aec2d0113562af2754219efbfc8cab5bb8d52ea063e21ef62f649f03705954ae8a6a763c7a1083495d74e0152bf6ca7f10e7e059a0177a17caba5d8e38b88dc64aec4766cce2f51f3a8f86c1190ea3400b4148c5a0a0b9d0fea8852f96d3a2b1c21b947be44bd3e8e69d0cc452a38bd0c50036f9236ba0748f73cec222e79796d9a051b3247a648b70174746116ccf2c6d44b7c5ca92dda0581393bfaa60b5ce83b2c08969b94197d8ce70900ee03d1944366ee9353236aba05bd17ccc2f8f82b53db52190ceb8e3938962411d02aac797a817d66366c214a6a0730033ce8b0c0da4eae4f9848e1a362f3f4dbe735a09e062ea5a115c4561cf9ca006b589e2210092bace9e9fcc63f6a7f564458049b4c0f0bfe06d10d1652710c6a0b2b3f246c4b195ddbc314f8844b501e50afd103b8b7568044a521e6f36f9a8eba0f91c990673b78bf5abc5538328089cf89c8302bf1904056f839dec7371a30c3ca0d1dd1ffed2a0b770bd03bd212b1bc20c6fd2eb2059c6e90d4dff6626730a27c3a08cb427dedc8ba43453ee3d4164ae1a155a9543c4c08db060d400b2158b999e34a06316d1af10873beb39a908e7c2766178f3cc431a9208e3f5f846f2b2fc20a760a0d81786e90402d275b01e281d522d140fda5b88a2f0786347e1b7aadafa03347fa0e93df0dfb320fd493a8f9ae3eb1083748fa7cd18e503b4d80f1e48a7a8d6611ea078165a0ea0547cd56dde852729cfdc14c8bddc3800304eb591ae5123e7b87de8a07af0fdb8183c5e3603b52038bf44f85677658f07bb7b99f9584f8d8596671ecba0daba8b73ce81af08b267bfe2d4319079f782dd5c7a58b727c800abeb1564c884a0eb8a0f0c09f3672f7840b08bf889bef596e2f55a63a365545d2d91ebd9b752f9a02dc9a3c2796ec2c87f65064f22d7f75ecaceb3a4a629f785cf1d62342b3392c8a0cf3a3fca53a0831c8fb90601d1ec6907299d8627d4b5debd3fd71ee795e11704a0beb3fc6f07969a7a0c72c8238607188ee4df3fb8b76123cd018bf64f976d38d4a085504c2f09cd7a8e080a20ca2e94b0c6c264b751170e7ec61b1b44a643e29186a0cd46e49571a79e8b76d48cafe0ecaba5e7ecbe82a1aca06a0623e9e8a21374c2a031c3934e927c3130fa162842ecdc82290e10cccddd93365b9b5065e7296f2e2d",
	}

	for i := range in {
		p := &PooledTransactionsRLPPacket{}
		err := rlp.DecodeBytes(common.FromHex(in[i]), p)
		assert.NoError(t, err)
	}
}
