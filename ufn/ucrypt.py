import base64
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC


class Crypt:
    def __init__(self, salt='HobbyMarks'):
        self.salt = salt.encode('UTF-8')
        self.kdf = PBKDF2HMAC(algorithm=hashes.SHA256(),
                              length=32,
                              salt=self.salt,
                              iterations=100000)

    def get_key(self, passwd):
        return base64.urlsafe_b64encode(self.kdf.derive(
            passwd.encode("UTF-8")))

    def encrypt(self, plaintext, passwd):
        key = self.get_key(passwd)
        if type(plaintext) is str:
            return Fernet(key).encrypt(plaintext.encode("UTF-8"))
        elif type(plaintext) is bytes:
            return Fernet(key).encrypt(plaintext)
        else:
            return None

    def decrypt(self, ciphertext, passwd):
        key = self.get_key(passwd)
        if type(ciphertext) is bytes:
            return Fernet(key).decrypt(ciphertext)
        elif type(ciphertext) is str:
            return Fernet(key).decrypt(ciphertext.encode("UTF-8"))
        else:
            return None


def encrypt_b64_str(plaintext="", passwd="", salt=""):
    if salt == "":
        crypt_ins = Crypt()
    else:
        crypt_ins = Crypt(salt)
    enc_bytes = crypt_ins.encrypt(plaintext, passwd)
    return base64.b64encode(enc_bytes).decode("UTF-8")


def b64_str_decrypt(ciphertext="", passwd="", salt=""):
    if salt == "":
        crypt_ins = Crypt()
    else:
        crypt_ins = Crypt(salt)
    enc_bytes = base64.b64decode(ciphertext)
    return crypt_ins.decrypt(enc_bytes, passwd).decode("UTF-8")
