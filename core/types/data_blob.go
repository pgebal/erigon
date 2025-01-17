package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/holiman/uint256"
	"github.com/protolambda/go-kzg/eth"
	"github.com/protolambda/ztyp/codec"

	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/hexutil"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/rlp"
)

// Compressed BLS12-381 G1 element
type KZGCommitment [48]byte

func (p *KZGCommitment) Deserialize(dr *codec.DecodingReader) error {
	if p == nil {
		return errors.New("nil pubkey")
	}
	_, err := dr.Read(p[:])
	return err
}

func (p *KZGCommitment) Serialize(w *codec.EncodingWriter) error {
	return w.Write(p[:])
}

func (KZGCommitment) ByteLength() uint64 {
	return 48
}

func (KZGCommitment) FixedLength() uint64 {
	return 48
}

func (p KZGCommitment) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p KZGCommitment) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *KZGCommitment) UnmarshalText(text []byte) error {
	return hexutil.UnmarshalFixedText("KZGCommitment", text, p[:])
}

func (c KZGCommitment) ComputeVersionedHash() common.Hash {
	return common.Hash(eth.KZGToVersionedHash(eth.KZGCommitment(c)))
}

// Compressed BLS12-381 G1 element
type KZGProof [48]byte

func (p *KZGProof) Deserialize(dr *codec.DecodingReader) error {
	if p == nil {
		return errors.New("nil pubkey")
	}
	_, err := dr.Read(p[:])
	return err
}

func (p *KZGProof) Serialize(w *codec.EncodingWriter) error {
	return w.Write(p[:])
}

func (KZGProof) ByteLength() uint64 {
	return 48
}

func (KZGProof) FixedLength() uint64 {
	return 48
}

func (p KZGProof) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p KZGProof) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *KZGProof) UnmarshalText(text []byte) error {
	return hexutil.UnmarshalFixedText("KZGProof", text, p[:])
}

// BLSFieldElement is the raw bytes representation of a field element
type BLSFieldElement [32]byte

func (p BLSFieldElement) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(p[:])), nil
}

func (p BLSFieldElement) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *BLSFieldElement) UnmarshalText(text []byte) error {
	return hexutil.UnmarshalFixedText("BLSFieldElement", text, p[:])
}

// Blob data
type Blob [params.FieldElementsPerBlob]BLSFieldElement

// eth.Blob interface
func (blob Blob) Len() int {
	return len(blob)
}

// eth.Blob interface
func (blob Blob) At(i int) [32]byte {
	return [32]byte(blob[i])
}

func (blob *Blob) Deserialize(dr *codec.DecodingReader) error {
	if blob == nil {
		return errors.New("cannot decode ssz into nil Blob")
	}
	for i := uint64(0); i < params.FieldElementsPerBlob; i++ {
		// TODO: do we want to check if each field element is within range?
		if _, err := dr.Read(blob[i][:]); err != nil {
			return err
		}
	}
	return nil
}

func (blob *Blob) Serialize(w *codec.EncodingWriter) error {
	for i := range blob {
		if err := w.Write(blob[i][:]); err != nil {
			return err
		}
	}
	return nil
}

func (blob *Blob) ByteLength() (out uint64) {
	return params.FieldElementsPerBlob * 32
}

func (blob *Blob) FixedLength() uint64 {
	return params.FieldElementsPerBlob * 32
}

func (blob *Blob) MarshalText() ([]byte, error) {
	out := make([]byte, 2+params.FieldElementsPerBlob*32*2)
	copy(out[:2], "0x")
	j := 2
	for _, elem := range blob {
		hex.Encode(out[j:j+64], elem[:])
		j += 64
	}
	return out, nil
}

func (blob *Blob) String() string {
	v, err := blob.MarshalText()
	if err != nil {
		return "<invalid-blob>"
	}
	return string(v)
}

func (blob *Blob) UnmarshalText(text []byte) error {
	if blob == nil {
		return errors.New("cannot decode text into nil Blob")
	}
	l := 2 + params.FieldElementsPerBlob*32*2
	if len(text) != l {
		return fmt.Errorf("expected %d characters but got %d", l, len(text))
	}
	if !(text[0] == '0' && text[1] == 'x') {
		return fmt.Errorf("expected '0x' prefix in Blob string")
	}
	j := 0
	for i := 2; i < l; i += 64 {
		if _, err := hex.Decode(blob[j][:], text[i:i+64]); err != nil {
			return fmt.Errorf("blob item %d is not formatted correctly: %v", j, err)
		}
		j += 1
	}
	return nil
}

type BlobKzgs []KZGCommitment

// eth.KZGCommitmentSequence interface
func (bk BlobKzgs) Len() int {
	return len(bk)
}

func (bk BlobKzgs) At(i int) eth.KZGCommitment {
	return eth.KZGCommitment(bk[i])
}

func (li *BlobKzgs) Deserialize(dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*li)
		*li = append(*li, KZGCommitment{})
		return &(*li)[i]
	}, 48, params.MaxBlobsPerBlock)
}

func (li BlobKzgs) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &li[i]
	}, 48, uint64(len(li)))
}

func (li BlobKzgs) ByteLength() uint64 {
	return uint64(len(li)) * 48
}

func (li BlobKzgs) FixedLength() uint64 {
	return 0
}

func (li BlobKzgs) copy() BlobKzgs {
	cpy := make(BlobKzgs, len(li))
	copy(cpy, li)
	return cpy
}

type Blobs []Blob

// eth.BlobSequence interface
func (blobs Blobs) Len() int {
	return len(blobs)
}

// eth.BlobSequence interface
func (blobs Blobs) At(i int) eth.Blob {
	return blobs[i]
}

func (a *Blobs) Deserialize(dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Blob{})
		return &(*a)[i]
	}, params.FieldElementsPerBlob*32, params.FieldElementsPerBlob)
}

func (a Blobs) Serialize(w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, params.FieldElementsPerBlob*32, uint64(len(a)))
}

func (a Blobs) ByteLength() (out uint64) {
	return uint64(len(a)) * params.FieldElementsPerBlob * 32
}

func (a *Blobs) FixedLength() uint64 {
	return 0 // it's a list, no fixed length
}

func (blobs Blobs) copy() Blobs {
	cpy := make(Blobs, len(blobs))
	copy(cpy, blobs) // each blob element is an array and gets deep-copied
	return cpy
}

// Return KZG commitments, versioned hashes and the aggregated KZG proof that correspond to these blobs
func (blobs Blobs) ComputeCommitmentsAndAggregatedProof() (commitments []KZGCommitment, versionedHashes []common.Hash, aggregatedProof KZGProof, err error) {
	commitments = make([]KZGCommitment, len(blobs))
	versionedHashes = make([]common.Hash, len(blobs))
	for i, blob := range blobs {
		c, ok := eth.BlobToKZGCommitment(blob)
		if !ok {
			return nil, nil, KZGProof{}, errors.New("could not convert blob to commitment")
		}
		commitments[i] = KZGCommitment(c)
		versionedHashes[i] = common.Hash(eth.KZGToVersionedHash(c))
	}

	var kzgProof KZGProof
	if len(blobs) != 0 {
		proof, err := eth.ComputeAggregateKZGProof(blobs)
		if err != nil {
			return nil, nil, KZGProof{}, err
		}
		kzgProof = KZGProof(proof)
	}

	return commitments, versionedHashes, kzgProof, nil
}

// BlobTxWrapper is the "network representation" of a Blob transaction, that is it includes not
// only the SignedBlobTx but also all the associated blob data.
type BlobTxWrapper struct {
	Tx                 SignedBlobTx
	BlobKzgs           BlobKzgs
	Blobs              Blobs
	KzgAggregatedProof KZGProof
}

func (txw *BlobTxWrapper) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&txw.Tx, &txw.BlobKzgs, &txw.Blobs, &txw.KzgAggregatedProof)
}

func (txw *BlobTxWrapper) Serialize(w *codec.EncodingWriter) error {
	return w.Container(&txw.Tx, &txw.BlobKzgs, &txw.Blobs, &txw.KzgAggregatedProof)
}

func (txw *BlobTxWrapper) ByteLength() uint64 {
	return codec.ContainerLength(&txw.Tx, &txw.BlobKzgs, &txw.Blobs, &txw.KzgAggregatedProof)
}

func (txw *BlobTxWrapper) FixedLength() uint64 {
	return 0
}

func (txw *BlobTxWrapper) VerifyBlobs() error {
	blobTx := txw.Tx.Message
	l1 := len(txw.BlobKzgs)
	l2 := len(blobTx.BlobVersionedHashes)
	l3 := len(txw.Blobs)
	if l1 != l2 || l2 != l3 {
		return fmt.Errorf("lengths don't match %v %v %v", l1, l2, l3)
	}
	// the following check isn't strictly necessary as it would be caught by data gas processing
	// (and hence it is not explicitly in the spec for this function), but it doesn't hurt to fail
	// early in case we are getting spammed with too many blobs or there is a bug somewhere:
	if l1 > params.MaxBlobsPerBlock {
		return fmt.Errorf("number of blobs exceeds max: %v", l1)
	}
	ok, err := eth.VerifyAggregateKZGProof(txw.Blobs, txw.BlobKzgs, eth.KZGProof(txw.KzgAggregatedProof))
	if err != nil {
		return fmt.Errorf("error during proof verification: %v", err)
	}
	if !ok {
		return errors.New("failed to verify kzg")
	}
	for i, h := range blobTx.BlobVersionedHashes {
		if computed := txw.BlobKzgs[i].ComputeVersionedHash(); computed != h {
			return fmt.Errorf("versioned hash %d supposedly %s but does not match computed %s", i, h, computed)
		}
	}
	return nil
}

// Implement transaction interface
func (txw *BlobTxWrapper) Type() byte               { return txw.Tx.Type() }
func (txw *BlobTxWrapper) GetChainID() *uint256.Int { return txw.Tx.GetChainID() }
func (txw *BlobTxWrapper) GetNonce() uint64         { return txw.Tx.GetNonce() }
func (txw *BlobTxWrapper) GetPrice() *uint256.Int   { return txw.Tx.GetPrice() }
func (txw *BlobTxWrapper) GetTip() *uint256.Int     { return txw.Tx.GetTip() }
func (txw *BlobTxWrapper) GetEffectiveGasTip(baseFee *uint256.Int) *uint256.Int {
	return txw.GetEffectiveGasTip(baseFee)
}
func (txw *BlobTxWrapper) GetFeeCap() *uint256.Int      { return txw.Tx.GetFeeCap() }
func (txw *BlobTxWrapper) Cost() *uint256.Int           { return txw.Tx.GetFeeCap() }
func (txw *BlobTxWrapper) GetDataHashes() []common.Hash { return txw.Tx.GetDataHashes() }
func (txw *BlobTxWrapper) GetGas() uint64               { return txw.Tx.GetGas() }
func (txw *BlobTxWrapper) GetDataGas() uint64           { return txw.Tx.GetDataGas() }
func (txw *BlobTxWrapper) GetValue() *uint256.Int       { return txw.Tx.GetValue() }
func (txw *BlobTxWrapper) Time() time.Time              { return txw.Tx.Time() }
func (txw *BlobTxWrapper) GetTo() *common.Address       { return txw.Tx.GetTo() }
func (txw *BlobTxWrapper) AsMessage(s Signer, baseFee *big.Int, rules *params.Rules) (Message, error) {
	return txw.Tx.AsMessage(s, baseFee, rules)
}
func (txw *BlobTxWrapper) WithSignature(signer Signer, sig []byte) (Transaction, error) {
	return txw.Tx.WithSignature(signer, sig)
}
func (txw *BlobTxWrapper) FakeSign(address common.Address) (Transaction, error) {
	return txw.Tx.FakeSign(address)
}
func (txw *BlobTxWrapper) Hash() common.Hash { return txw.Tx.Hash() }
func (txw *BlobTxWrapper) SigningHash(chainID *big.Int) common.Hash {
	return txw.Tx.SigningHash(chainID)
}
func (txw *BlobTxWrapper) GetData() []byte           { return txw.Tx.GetData() }
func (txw *BlobTxWrapper) GetAccessList() AccessList { return txw.Tx.GetAccessList() }
func (txw *BlobTxWrapper) Protected() bool           { return txw.Tx.Protected() }
func (txw *BlobTxWrapper) RawSignatureValues() (*uint256.Int, *uint256.Int, *uint256.Int) {
	return txw.Tx.RawSignatureValues()
}
func (txw *BlobTxWrapper) Sender(s Signer) (common.Address, error) { return txw.Tx.Sender(s) }
func (txw *BlobTxWrapper) GetSender() (common.Address, bool)       { return txw.Tx.GetSender() }
func (txw *BlobTxWrapper) SetSender(address common.Address)        { txw.Tx.SetSender(address) }
func (txw *BlobTxWrapper) IsContractDeploy() bool                  { return txw.Tx.IsContractDeploy() }
func (txw *BlobTxWrapper) IsStarkNet() bool                        { return false }

func (txw *BlobTxWrapper) Size() common.StorageSize {
	if size := txw.Tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := txw.EncodingSize()
	txw.Tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

func (txw BlobTxWrapper) EncodingSize() int {
	envelopeSize := int(codec.ContainerLength(&txw.Tx, &txw.BlobKzgs, &txw.Blobs, &txw.KzgAggregatedProof))
	// Add type byte
	envelopeSize++
	return envelopeSize
}

func (txw *BlobTxWrapper) MarshalBinary(w io.Writer) error {
	var b [33]byte
	// encode TxType
	b[0] = BlobTxType
	if _, err := w.Write(b[:1]); err != nil {
		return err
	}
	wcodec := codec.NewEncodingWriter(w)
	return txw.Serialize(wcodec)
}

func (txw BlobTxWrapper) EncodeRLP(w io.Writer) error {
	var buf bytes.Buffer
	if err := txw.MarshalBinary(&buf); err != nil {
		return err
	}
	return rlp.Encode(w, buf.Bytes())
}
