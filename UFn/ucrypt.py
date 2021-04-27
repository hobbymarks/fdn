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

    def get_key(self, password_str):
        """

        :param password_str:
        :type password_str:
        :return:
        :rtype:
        """
        return base64.urlsafe_b64encode(
            self.kdf.derive(password_str.encode("UTF-8")))

    def encrypt(self, str_to_enc, password_str):
        """

        :param str_to_enc:
        :type str_to_enc:
        :param password_str:
        :type password_str:
        :return:
        :rtype:
        """
        key = self.get_key(password_str)
        if type(str_to_enc) is str:
            return Fernet(key).encrypt(str_to_enc.encode("UTF-8"))
        elif type(str_to_enc) is bytes:
            return Fernet(key).encrypt(str_to_enc)
        else:
            return None

    def decrypt(self, enc_str, password_str):
        """

        :param enc_str:
        :type enc_str:
        :param password_str:
        :type password_str:
        :return:
        :rtype:
        """
        key = key = self.get_key(password_str)
        if type(enc_str) is bytes:
            return Fernet(key).decrypt(enc_str)
        elif type(enc_str) is str:
            return Fernet(key).decrypt(enc_str.encode("UTF-8"))
        else:
            return None


def encrypt_b64_str(str_to_enc="", password_str="", salt_str=""):
    """

    :param str_to_enc:
    :type str_to_enc:
    :param password_str:
    :type password_str:
    :param salt_str:
    :type salt_str:
    :return:
    :rtype:
    """
    if salt_str == "":
        crypt_ins = Crypt()
    else:
        crypt_ins = Crypt(salt_str)
    enc_bytes = crypt_ins.encrypt(str_to_enc, password_str)
    return base64.b64encode(enc_bytes).decode("UTF-8")


def b64_str_decrypt(enc_str="", password_str="", salt_str=""):
    """

    :param enc_str:
    :type enc_str:
    :param password_str:
    :type password_str:
    :param salt_str:
    :type salt_str:
    :return:
    :rtype:
    """
    if salt_str == "":
        crypt_ins = Crypt()
    else:
        crypt_ins = Crypt(salt_str)
    enc_bytes = base64.b64decode(enc_str)
    return crypt_ins.decrypt(enc_bytes, password_str).decode("UTF-8")
