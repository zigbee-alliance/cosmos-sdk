package types

import (
	ics23 "github.com/confio/ics23/go"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ clientexported.ClientState = ClientState{}

// ClientState of a Solo Machine represents whether or not the client is frozen.
type ClientState struct {
	// Client ID
	ID string `json:"id" yaml:"id"`

	ChainID string `json:"chain_id" yaml:"chain_id"`

	// Frozen status of the client
	Frozen bool `json:"frozen" yaml:"frozen"`

	// Current consensus state of the client
	ConsensusState ConsensusState `json:"consensus_state" yaml:"consensus_state"`
}

// NewClientState creates a new ClientState instance.
func NewClientState(id, chainID string, consensusState ConsensusState) ClientState {
	return ClientState{
		ID:             id,
		ChainID:        chainID,
		Frozen:         false,
		ConsensusState: consensusState,
	}
}

// GetID returns the solo machine client state identifier.
func (cs ClientState) GetID() string {
	return cs.ID
}

// GetChainID returns an empty string.
func (cs ClientState) GetChainID() string {
	return cs.ChainID
}

// ClientType is Solo Machine.
func (cs ClientState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetLatestHeight returns the latest sequence number.
func (cs ClientState) GetLatestHeight() uint64 {
	return cs.ConsensusState.Sequence
}

// IsFrozen returns true if the client is frozen
func (cs ClientState) IsFrozen() bool {
	return cs.Frozen
}

// GetProofSpecs returns nil proof specs since client state verification uses signatures.
func (cs ClientState) GetProofSpecs() []*ics23.ProofSpec {
	return nil
}

// Validate performs basic validation of the client state fields.
func (cs ClientState) Validate() error {
	if err := host.ClientIdentifierValidator(cs.ID); err != nil {
		return err
	}
	return cs.ConsensusState.ValidateBasic()
}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// Solo Machine client stored on the target machine.
func (cs ClientState) VerifyClientConsensusState(
	store sdk.KVStore,
	cdc codec.Marshaler,
	aminoCdc *codec.Codec,
	root commitmentexported.Root,
	sequence uint64,
	counterpartyClientIdentifier string,
	consensusHeight uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	_ clientexported.ConsensusState,
) error {
	signature, err := sanitizeVerificationArgs(cdc, cs, sequence, prefix, proof, cs.ConsensusState)
	if err != nil {
		return err
	}

	clientPrefixedPath := "clients/" + counterpartyClientIdentifier + "/" + host.ConsensusStatePath(consensusHeight)
	path, err := commitmenttypes.ApplyPrefix(prefix, clientPrefixedPath)
	if err != nil {
		return err
	}

	data, err := ConsensusStateSignBytes(aminoCdc, sequence, signature.Timestamp, path, cs.ConsensusState)
	if err != nil {
		return err
	}

	if err := CheckSignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedClientConsensusStateVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cs)
	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (cs ClientState) VerifyConnectionState(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	_ clientexported.ConsensusState,
) error {
	signature, err := sanitizeVerificationArgs(cdc, cs, sequence, prefix, proof, cs.ConsensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	data, err := ConnectionStateSignBytes(cdc, sequence, signature.Timestamp, path, connectionEnd)
	if err != nil {
		return err
	}

	if err := CheckSignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedConnectionStateVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cs)
	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (cs ClientState) VerifyChannelState(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	_ clientexported.ConsensusState,
) error {
	signature, err := sanitizeVerificationArgs(cdc, cs, sequence, prefix, proof, cs.ConsensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	data, err := ChannelStateSignBytes(cdc, sequence, signature.Timestamp, path, channel)
	if err != nil {
		return err
	}

	if err := CheckSignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedChannelStateVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cs)
	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketCommitment(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	commitmentBytes []byte,
	_ clientexported.ConsensusState,
) error {
	signature, err := sanitizeVerificationArgs(cdc, cs, sequence, prefix, proof, cs.ConsensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketCommitmentPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketCommitmentSignBytes(sequence, signature.Timestamp, path, commitmentBytes)

	if err := CheckSignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketCommitmentVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cs)
	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (cs ClientState) VerifyPacketAcknowledgement(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	acknowledgement []byte,
	_ clientexported.ConsensusState,
) error {
	signature, err := sanitizeVerificationArgs(cdc, cs, sequence, prefix, proof, cs.ConsensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketAcknowledgementSignBytes(sequence, signature.Timestamp, path, acknowledgement)

	if err := CheckSignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cs)
	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (cs ClientState) VerifyPacketAcknowledgementAbsence(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	packetSequence uint64,
	_ clientexported.ConsensusState,
) error {
	signature, err := sanitizeVerificationArgs(cdc, cs, sequence, prefix, proof, cs.ConsensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.PacketAcknowledgementPath(portID, channelID, packetSequence))
	if err != nil {
		return err
	}

	data := PacketAcknowledgementAbsenceSignBytes(sequence, signature.Timestamp, path)

	if err := CheckSignature(cs.ConsensusState.GetPubKey(), data, signature.Signature); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckAbsenceVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cs)
	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (cs ClientState) VerifyNextSequenceRecv(
	store sdk.KVStore,
	cdc codec.Marshaler,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	_ clientexported.ConsensusState,
) error {
	signature, err := sanitizeVerificationArgs(cdc, cs, sequence, prefix, proof, cs.ConsensusState)
	if err != nil {
		return err
	}

	path, err := commitmenttypes.ApplyPrefix(prefix, host.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	data := NextSequenceRecvSignBytes(sequence, signature.Timestamp, path, nextSequenceRecv)

	if err := CheckSignature(cs.ConsensusState.GetPubKey(), data, proof); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrFailedNextSeqRecvVerification, err.Error())
	}

	cs.ConsensusState.Sequence++
	cs.ConsensusState.Timestamp = signature.Timestamp
	setClientState(store, cs)
	return nil
}

// sanitizeVerificationArgs perfoms the basic checks on the arguments that are
// shared between the verification functions and returns the unmarshalled
// proof representing the signature and timestamp.
func sanitizeVerificationArgs(
	cdc codec.Marshaler,
	cs ClientState,
	sequence uint64,
	prefix commitmentexported.Prefix,
	proof []byte,
	consensusState clientexported.ConsensusState,
) (signature Signature, err error) {
	if cs.GetLatestHeight() < sequence {
		return Signature{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"client state (%s) sequence < proof sequence (%d < %d)", cs.ID, cs.GetLatestHeight(), sequence,
		)
	}

	if cs.IsFrozen() {
		return Signature{}, clienttypes.ErrClientFrozen
	}

	if prefix == nil {
		return Signature{}, sdkerrors.Wrap(commitmenttypes.ErrInvalidPrefix, "prefix cannot be empty")
	}

	_, ok := prefix.(commitmenttypes.MerklePrefix)
	if !ok {
		return Signature{}, sdkerrors.Wrapf(commitmenttypes.ErrInvalidPrefix, "invalid prefix type %T, expected MerklePrefix", prefix)
	}

	if proof == nil {
		return Signature{}, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof cannot be empty")
	}

	if err = cdc.UnmarshalBinaryBare(proof, &signature); err != nil {
		return Signature{}, sdkerrors.Wrapf(ErrInvalidProof, "failed to unmarshal proof into type %T", Signature{})
	}

	if consensusState == nil {
		return Signature{}, sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be empty")
	}

	_, ok = consensusState.(ConsensusState)
	if !ok {
		return Signature{}, sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "invalid consensus type %T, expected %T", consensusState, ConsensusState{})
	}

	if consensusState.GetTimestamp() > signature.Timestamp {
		return Signature{}, sdkerrors.Wrapf(ErrInvalidProof, "the timestamp of the signature must be greater than or equal to the timestamp of the consensus state (%d >= %d)", signature.Timestamp, consensusState.GetTimestamp())
	}

	return signature, nil
}

// sets the client state to the store
func setClientState(store sdk.KVStore, clientState clientexported.ClientState) {
	bz := amino.MustMarshalBinaryBare(clientState)
	store.Set(host.KeyClientState(), bz)
}