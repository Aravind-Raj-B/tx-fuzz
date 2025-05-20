from eth_keyfile import load_keyfile, decode_keyfile_json

keystore_path = "/home/aravindh/Documents/common-test/data/execution-data-1/keystore/UTC--2025-05-19T13-17-56.996033077Z--46367479e1592f342a9b0fbbcd57eea2fad15b1d"
password = "1234567890"

# Open as a text file, not binary
with open(keystore_path, "r", encoding="utf-8") as f:
    keyfile_json = load_keyfile(f)
    private_key = decode_keyfile_json(keyfile_json, password.encode())

print("Private key (hex):", private_key.hex())
