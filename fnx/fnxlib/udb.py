import sqlite3
from datetime import datetime
from pathlib import Path

# From Third Party
import pandas as pd


class UDB:
    def __init__(self, db_path: Path):
        self._db_path = db_path
        self._db_con = sqlite3.connect(self._db_path)
        self._db_cur = self._db_con.cursor()
        self._tb_name = "fnxrd"

    def create_tb(self, tb_name: str):
        if tb_name:
            self._tb_name = tb_name
        tb_create_sql = """
        CREATE TABLE {} (
        id integer PRIMARY KEY,
        newID text NOT NULL,
        curID text NOT NULL,
        curCrypt text NOT NULL,
        sepDirPathID text NOT NULL,
        absDirPathID text NOT NULL,
        opStamp timestamp);
        """.format(self._tb_name)
        if not self.is_tb_exist(self._tb_name):
            self._db_cur.execute(tb_create_sql)

    # def change_tb_name(self,pre_name:str,new_name:str):
    #     tb_name_sql="""
    #
    #     """

    def insert_rd(self, new_id: str, cur_id: str, cur_crypt: str,
                  sep_dp_id: str, abs_dp_id: str):
        if not self.is_tb_exist(self._tb_name):
            self.create_tb(self._tb_name)
        insert_rd_sql = """
        INSERT INTO {} (
        'newID', 
        'curID', 
        'curCrypt',
        'sepDirPathID',
        'absDirPathID', 
        'opStamp') 
        VALUES (?,?,?,?,?,?);
        """.format(self._tb_name)
        self._db_cur.execute(
            insert_rd_sql,
            (new_id, cur_id, cur_crypt, sep_dp_id, abs_dp_id, datetime.now()))

    def checkout_rd(self, new_id: str):
        checkout_rd_sql = """
        SELECT curCrypt, opStamp FROM {} 
        WHERE newID=? ORDER BY opStamp DESC;
        """.format(self._tb_name)
        try:
            self._db_cur.execute(checkout_rd_sql, (new_id,))
            rows = self._db_cur.fetchall()
            return pd.DataFrame(rows, columns=["curCrypt", "opStamp"])
        except sqlite3.OperationalError as e:  # TODO: hidden error !!!
            return pd.DataFrame([], columns=["curCrypt", "opStamp"])

    def is_tb_exist(self, tb_name: str):
        tb_slt_sql = """
        SELECT count(name) FROM sqlite_master
        WHERE type='table' 
        AND name=?;
        """
        self._db_cur.execute(tb_slt_sql, (tb_name,))
        if self._db_cur.fetchone()[0] == 1:
            return True
        return False

    def commit(self):
        self._db_con.commit()

    def close(self):
        self._db_con.commit()
        self._db_con.close()
