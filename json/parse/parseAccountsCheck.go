package scenjsonparse

import (
	"errors"
	"fmt"

	mj "github.com/bhagyaraj1208117/andes-scenario-go/model"
	oj "github.com/bhagyaraj1208117/andes-scenario-go/orderedjson"
)

func (p *Parser) processCheckAccount(acctRaw oj.OJsonObject) (*mj.CheckAccount, error) {
	acctMap, isMap := acctRaw.(*oj.OJsonMap)
	if !isMap {
		return nil, errors.New("unmarshalled account object is not a map")
	}

	acct := mj.CheckAccount{
		Comment:              "",
		Nonce:                mj.JSONCheckUint64Unspecified(),
		Balance:              mj.JSONCheckBigIntUnspecified(),
		Username:             mj.JSONCheckBytesUnspecified(),
		ExplicitStorage:      false,
		IgnoreStorage:        true,
		MoreStorageAllowed:   false,
		CheckStorage:         nil,
		Code:                 mj.JSONCheckBytesUnspecified(),
		CodeMetadata:         mj.JSONCheckBytesUnspecified(),
		Owner:                mj.JSONCheckBytesUnspecified(),
		AsyncCallData:        mj.JSONCheckBytesUnspecified(),
		IgnoreDCT:            false,
		MoreDCTTokensAllowed: false,
		CheckDCTData:         nil,
		DeveloperReward:      mj.JSONCheckBigIntUnspecified(),
	}
	var err error

	for _, kvp := range acctMap.OrderedKV {
		switch kvp.Key {
		case "comment":
			acct.Comment, err = p.parseString(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid check account comment: %w", err)
			}
		case "nonce":
			acct.Nonce, err = p.processCheckUint64(kvp.Value)
			if err != nil {
				return nil, errors.New("invalid account nonce")
			}
		case "balance":
			acct.Balance, err = p.processCheckBigInt(kvp.Value, bigIntUnsignedBytes)
			if err != nil {
				return nil, errors.New("invalid account balance")
			}
		case "dct":
			acct.IgnoreDCT = IsStar(kvp.Value)
			if !acct.IgnoreDCT {
				dctMap, dctOk := kvp.Value.(*oj.OJsonMap)
				if !dctOk {
					return nil, errors.New("invalid DCT map")
				}
				for _, dctKvp := range dctMap.OrderedKV {
					if dctKvp.Key == "+" {
						acct.MoreDCTTokensAllowed = true
					} else {
						tokenNameStr, err := p.ExprInterpreter.InterpretString(dctKvp.Key)
						if err != nil {
							return nil, fmt.Errorf("invalid dct token identifer: %w", err)
						}
						tokenName := mj.NewJSONBytesFromString(tokenNameStr, dctKvp.Key)
						dctItem, err := p.processCheckDCTData(tokenName, dctKvp.Value)
						if err != nil {
							return nil, fmt.Errorf("invalid dct value: %w", err)
						}
						acct.CheckDCTData = append(acct.CheckDCTData, dctItem)
					}
				}
			}
		case "username":
			acct.Username, err = p.parseCheckBytes(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account username: %w", err)
			}
		case "storage":
			acct.ExplicitStorage = true
			acct.IgnoreStorage = IsStar(kvp.Value)
			if !acct.IgnoreStorage {
				storageMap, storageOk := kvp.Value.(*oj.OJsonMap)
				if !storageOk {
					return nil, errors.New("invalid account storage")
				}
				for _, storageKvp := range storageMap.OrderedKV {
					if storageKvp.Key == "+" {
						acct.MoreStorageAllowed = true
					} else {
						byteKey, err := p.ExprInterpreter.InterpretString(storageKvp.Key)
						if err != nil {
							return nil, fmt.Errorf("invalid account storage key: %w", err)
						}
						byteVal, err := p.parseCheckBytes(storageKvp.Value)
						if err != nil {
							return nil, fmt.Errorf("invalid account storage value: %w", err)
						}
						stElem := mj.CheckStorageKeyValuePair{
							Key:        mj.NewJSONBytesFromString(byteKey, storageKvp.Key),
							CheckValue: byteVal,
						}
						acct.CheckStorage = append(acct.CheckStorage, &stElem)
					}
				}
			}
		case "code":
			acct.Code, err = p.parseCheckBytes(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account code: %w", err)
			}
		case "codeMetadata":
			acct.CodeMetadata, err = p.parseCheckBytes(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account codeMetadata: %w", err)
			}
		case "owner":
			acct.Owner, err = p.parseCheckBytes(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid account owner: %w", err)
			}
		case "asyncCallData":
			acct.AsyncCallData, err = p.parseCheckBytes(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid asyncCallData: %w", err)
			}
		case "developerRewards":
			acct.DeveloperReward, err = p.processCheckBigInt(kvp.Value, bigIntUnsignedBytes)
			if err != nil {
				return nil, fmt.Errorf("invalid developerRewards: %w", err)
			}

		default:
			return nil, fmt.Errorf("unknown account field: %s", kvp.Key)
		}
	}

	return &acct, nil
}

func (p *Parser) processCheckAccountMap(acctMapRaw oj.OJsonObject) (*mj.CheckAccounts, error) {
	var checkAccounts = &mj.CheckAccounts{
		Accounts:            nil,
		MoreAccountsAllowed: false,
	}

	preMap, isPreMap := acctMapRaw.(*oj.OJsonMap)
	if !isPreMap {
		return nil, errors.New("unmarshalled check account map object is not a map")
	}
	for _, acctKVP := range preMap.OrderedKV {
		if acctKVP.Key == "+" {
			checkAccounts.MoreAccountsAllowed = true
		} else {
			acct, acctErr := p.processCheckAccount(acctKVP.Value)
			if acctErr != nil {
				return nil, acctErr
			}
			acctAddr, hexErr := p.parseAccountAddress(acctKVP.Key)
			if hexErr != nil {
				return nil, hexErr
			}
			acct.Address = acctAddr
			checkAccounts.Accounts = append(checkAccounts.Accounts, acct)
		}
	}
	return checkAccounts, nil
}
