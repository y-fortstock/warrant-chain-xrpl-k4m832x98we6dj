		// mnemonic := "hold gesture void wage high peasant sketch firm banner dragon judge hover"

	// // РАБОТАЮЩИЙ ЭТАЛОН
	// seed2, err := keypairs.GenerateSeed(mnemonic, ac.ED25519)
	// assert.NoError(t, err)
	// fmt.Println("seed2: ", seed2)

	// priv2, pub2, err := keypairs.DeriveKeypair(seed2, false)
	// assert.NoError(t, err)
	// fmt.Println("priv2: ", priv2)
	// fmt.Println("pub2: ", pub2)


  func TestGetKeyPairFromSeed_1(t *testing.T) {
	familySeed := "pNURfEJaBcFR15a1X4Zb6sJKuezyuVHZF5XVhTM9uFSCsyUw8WkRu"

	priv, pub, err := keypairs.DeriveKeypair(familySeed, false)
	assert.NoError(t, err)
	fmt.Println("priv: ", priv)
	fmt.Println("pub: ", pub)

	// Получаем адрес из публичного ключа
	pubKeyBytes, err := hex.DecodeString(pub)
	assert.NoError(t, err)

	// Используем правильный способ генерации адреса XRPL
	accountID := ac.Sha256RipeMD160(pubKeyBytes)
	accountAddress := ac.Encode(accountID, []byte{ac.AccountAddressPrefix}, ac.AccountAddressLength)
	fmt.Println("accountAddress: ", accountAddress)

	// Получаем текущий sequence number для аккаунта
	rpcCfg, err := client.NewJsonRpcConfig("https://s.altnet.rippletest.net:51234", client.WithHttpClient(&http.Client{
		Timeout: time.Duration(30) * time.Second,
	}))
	assert.NoError(t, err)

	cli := jsonrpcclient.NewClient(rpcCfg)

	// Получаем информацию об аккаунте
	accountInfoReq := &clientaccount.AccountInfoRequest{
		Account: types.Address(accountAddress),
	}
	accountInfo, _, err := cli.Account.AccountInfo(accountInfoReq)
	var sequence uint32 = 1
	if err != nil {
		fmt.Printf("Warning: Could not get account info for %s: %v\n", accountAddress, err)
		fmt.Println("This might be a new account that needs funding")
		// Для нового аккаунта используем sequence = 1
	} else {
		sequence = accountInfo.AccountData.Sequence
	}

	// Получаем текущий ledger
	ledgerReq := &clientledger.LedgerRequest{
		LedgerIndex: clientcommon.VALIDATED,
	}
	ledgerResp, _, err := cli.Ledger.Ledger(ledgerReq)
	assert.NoError(t, err)

	// Конвертируем LedgerIndex в uint32
	ledgerIndex := uint32(ledgerResp.LedgerIndex) + 20

	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account:            types.Address(accountAddress),
			TransactionType:    transactions.PaymentTx,
			Fee:                types.XRPCurrencyAmount(12000), // Увеличиваем fee
			Sequence:           sequence,
			LastLedgerSequence: ledgerIndex, // Добавляем LastLedgerSequence
			SigningPubKey:      pub,         // Добавляем публичный ключ для подписи
		},
		Amount:      types.XRPCurrencyAmount(1000000),
		Destination: types.Address("ra5nK24KXen9AHvsdFTKHSANinZseWnPcX"),
	}
	encodedForSigning, err := binarycodec.EncodeForSigning(payment)
	assert.NoError(t, err)
	fmt.Println("encodedForSigning: ", encodedForSigning)

	signature, err := keypairs.Sign(encodedForSigning, priv)
	assert.NoError(t, err)
	fmt.Println("signature: ", signature)

	payment.TxnSignature = signature

	txBlob, err := binarycodec.Encode(payment)
	assert.NoError(t, err)
	fmt.Println("txBlob: ", txBlob)

	submitReq := &clienttransactions.SubmitRequest{
		TxBlob: txBlob,
	}

	resp, xrplResp, err := cli.Transaction.Submit(submitReq)
	if err != nil {
		fmt.Printf("Submit error: %v\n", err)
		if xrplResp != nil {
			fmt.Printf("XRPL Response: %+v\n", xrplResp)
		}
		// Не делаем assert.NoError здесь, так как аккаунт может не иметь средств
		return
	}
	fmt.Println("resp: ", resp)
	fmt.Println("xrplResp: ", xrplResp)
}
  
  // КОНЕЦ РАБОТАЮЩЕГО ЭТАЛОНА

	// НУЖНО ДОВЕСТИ ДО РАБОЧЕГО ВИДА
	// Генерация seed (массив байт)
	// seed := bip39.NewSeed(mnemonic, "password") // НЕ МЕНЯТЬ
	// // Перевод seed в hex-строку
	// hexSeed := hex.EncodeToString(seed) // НЕ МЕНЯТЬ
	// fmt.Println("hexSeed: ", hexSeed)   // НЕ МЕНЯТЬ

	// seed, err := hex.DecodeString(hexSeed)
	// assert.NoError(t, err)

	// // 2. Create master key using btcsuite
	// masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	// assert.NoError(t, err)

	// // 3. Derive using XRP path m/44'/144'/0'/0/0
	// purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	// assert.NoError(t, err)

	// coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + 144)
	// assert.NoError(t, err)

	// account, err := coinType.Derive(hdkeychain.HardenedKeyStart + 0)
	// assert.NoError(t, err)

	// change, err := account.Derive(0)
	// assert.NoError(t, err)

	// addressIndex, err := change.Derive(0)
	// assert.NoError(t, err)

	// // 4. Extract private key (32 bytes)
	// privateKey, err := addressIndex.ECPrivKey()
	// assert.NoError(t, err)
	// privateKeyBytes := privateKey.Serialize()

	// // 5. Take first 16 bytes for XRPL entropy
	// entropy := privateKeyBytes[:16]

	// // 6. Encode as XRPL family seed
	// // This requires XRPL's addresscodec.EncodeSeed() equivalent
	// familySeed, err := EncodeXRPLSeed(entropy, ED25519_PREFIX)
	// fmt.Println("familySeed: ", familySeed)

	// ================================
	// familySeed := "pNURfEJaBcFR15a1X4Zb6sJKuezyuVHZF5XVhTM9uFSCsyUw8WkRu"

	// priv, pub, err := keypairs.DeriveKeypair(familySeed, false)
	// assert.NoError(t, err)
	// fmt.Println("priv: ", priv)
	// fmt.Println("pub: ", pub)

	// actual, err := keypairs.Sign([]byte("test"), priv)
	// assert.NoError(t, err)
	// fmt.Println("actual: ", actual)