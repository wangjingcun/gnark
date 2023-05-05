// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark DO NOT EDIT

package plonk

import (
	curve "github.com/consensys/gnark-crypto/ecc/bn254"

	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/iop"
	"io"
)

// WriteRawTo writes binary encoding of Proof to w without point compression
func (proof *Proof) WriteRawTo(w io.Writer) (int64, error) {
	return proof.writeTo(w, curve.RawEncoding())
}

// WriteTo writes binary encoding of Proof to w with point compression
func (proof *Proof) WriteTo(w io.Writer) (int64, error) {
	return proof.writeTo(w)
}

func (proof *Proof) writeTo(w io.Writer, options ...func(*curve.Encoder)) (int64, error) {
	enc := curve.NewEncoder(w, options...)

	toEncode := []interface{}{
		&proof.LRO[0],
		&proof.LRO[1],
		&proof.LRO[2],
		&proof.Z,
		&proof.H[0],
		&proof.H[1],
		&proof.H[2],
		&proof.BatchedProof.H,
		proof.BatchedProof.ClaimedValues,
		&proof.ZShiftedOpening.H,
		&proof.ZShiftedOpening.ClaimedValue,
		&proof.PI2,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom reads binary representation of Proof from r
func (proof *Proof) ReadFrom(r io.Reader) (int64, error) {
	dec := curve.NewDecoder(r)
	toDecode := []interface{}{
		&proof.LRO[0],
		&proof.LRO[1],
		&proof.LRO[2],
		&proof.Z,
		&proof.H[0],
		&proof.H[1],
		&proof.H[2],
		&proof.BatchedProof.H,
		&proof.BatchedProof.ClaimedValues,
		&proof.ZShiftedOpening.H,
		&proof.ZShiftedOpening.ClaimedValue,
		&proof.PI2,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of ProvingKey to w
func (pk *ProvingKey) WriteTo(w io.Writer) (n int64, err error) {
	// encode the verifying key
	n, err = pk.Vk.WriteTo(w)
	if err != nil {
		return
	}

	// fft domains
	n2, err := pk.Domain[0].WriteTo(w)
	if err != nil {
		return
	}
	n += n2

	n2, err = pk.Domain[1].WriteTo(w)
	if err != nil {
		return
	}
	n += n2

	// sanity check len(Permutation) == 3*int(pk.Domain[0].Cardinality)
	if len(pk.trace.S) != (3 * int(pk.Domain[0].Cardinality)) {
		return n, errors.New("invalid permutation size, expected 3*domain cardinality")
	}

	enc := curve.NewEncoder(w)
	// note: type Polynomial, which is handled by default binary.Write(...) op and doesn't
	// encode the size (nor does it convert from Montgomery to Regular form)
	// so we explicitly transmit []fr.Element
	toEncode := []interface{}{
		pk.trace.Ql.Coefficients(),
		pk.trace.Qr.Coefficients(),
		pk.trace.Qm.Coefficients(),
		pk.trace.Qo.Coefficients(),
		pk.trace.Qk.Coefficients(),
		coefficients(pk.trace.Qcp),
		pk.lQk.Coefficients(),
		pk.trace.S1.Coefficients(),
		pk.trace.S2.Coefficients(),
		pk.trace.S3.Coefficients(),
		pk.trace.S,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return n + enc.BytesWritten(), err
		}
	}

	return n + enc.BytesWritten(), nil
}

// ReadFrom reads from binary representation in r into ProvingKey
func (pk *ProvingKey) ReadFrom(r io.Reader) (int64, error) {
	pk.Vk = &VerifyingKey{}
	n, err := pk.Vk.ReadFrom(r)
	if err != nil {
		return n, err
	}

	n2, err := pk.Domain[0].ReadFrom(r)
	n += n2
	if err != nil {
		return n, err
	}

	n2, err = pk.Domain[1].ReadFrom(r)
	n += n2
	if err != nil {
		return n, err
	}

	pk.trace.S = make([]int64, 3*pk.Domain[0].Cardinality)

	dec := curve.NewDecoder(r)

	var ql, qr, qm, qo, qk, lqk, s1, s2, s3 []fr.Element
	var qcp [][]fr.Element
	toDecode := []interface{}{
		&ql,
		&qr,
		&qm,
		&qo,
		&qk,
		&qcp,
		&lqk,
		&s1,
		&s2,
		&s3,
		&pk.trace.S,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return n + dec.BytesRead(), err
		}
	}

	canReg := iop.Form{Basis: iop.Canonical, Layout: iop.Regular}
	pk.trace.Ql = iop.NewPolynomial(&ql, canReg)
	pk.trace.Qr = iop.NewPolynomial(&qr, canReg)
	pk.trace.Qm = iop.NewPolynomial(&qm, canReg)
	pk.trace.Qo = iop.NewPolynomial(&qo, canReg)
	pk.trace.Qk = iop.NewPolynomial(&qk, canReg)
	pk.trace.S1 = iop.NewPolynomial(&s1, canReg)
	pk.trace.S2 = iop.NewPolynomial(&s2, canReg)
	pk.trace.S3 = iop.NewPolynomial(&s3, canReg)

	pk.trace.Qcp = make([]*iop.Polynomial, len(qcp))
	for i := range qcp {
		pk.trace.Qcp[i] = iop.NewPolynomial(&qcp[i], canReg)
	}

	lagReg := iop.Form{Basis: iop.Lagrange, Layout: iop.Regular}
	pk.lQk = iop.NewPolynomial(&lqk, lagReg)

	pk.computeLagrangeCosetPolys()

	return n + dec.BytesRead(), nil

}

// WriteTo writes binary encoding of VerifyingKey to w
func (vk *VerifyingKey) WriteTo(w io.Writer) (n int64, err error) {
	return vk.writeTo(w)
}

// WriteRawTo writes binary encoding of VerifyingKey to w without point compression
func (vk *VerifyingKey) WriteRawTo(w io.Writer) (int64, error) {
	return vk.writeTo(w, curve.RawEncoding())
}

func (vk *VerifyingKey) writeTo(w io.Writer, options ...func(*curve.Encoder)) (n int64, err error) {
	enc := curve.NewEncoder(w)

	toEncode := []interface{}{
		vk.Size,
		&vk.SizeInv,
		&vk.Generator,
		vk.NbPublicVariables,
		&vk.CosetShift,
		&vk.S[0],
		&vk.S[1],
		&vk.S[2],
		&vk.Ql,
		&vk.Qr,
		&vk.Qm,
		&vk.Qo,
		&vk.Qk,
		&vk.Qcp,
		vk.CommitmentConstraintIndexes,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}

// ReadFrom reads from binary representation in r into VerifyingKey
func (vk *VerifyingKey) ReadFrom(r io.Reader) (int64, error) {
	dec := curve.NewDecoder(r)
	toDecode := []interface{}{
		&vk.Size,
		&vk.SizeInv,
		&vk.Generator,
		&vk.NbPublicVariables,
		&vk.CosetShift,
		&vk.S[0],
		&vk.S[1],
		&vk.S[2],
		&vk.Ql,
		&vk.Qr,
		&vk.Qm,
		&vk.Qo,
		&vk.Qk,
		&vk.Qcp,
		&vk.CommitmentConstraintIndexes,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}
