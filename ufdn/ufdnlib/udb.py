from datetime import datetime
import os
import sqlite3

# From Third Party
import pandas as pd

# From This Project
from ufdn.ufdnlib import uconfig


class UDB:
    def __init__(self, db_path=""):
        if not db_path:
            self._db_path = os.path.join(uconfig.gParamDict["record_path"],
                                         "rd.db")
        else:
            self._db_path = db_path
        self._db_con = sqlite3.connect(self._db_path)
        self._db_cur = self._db_con.cursor()
        self._tb_name = "UFnRecord"

    def create_tb(self, tb_name):
        if tb_name and (type(tb_name) is str):
            self._tb_name = tb_name
        tb_create_sql = """
        CREATE TABLE {} (
        id integer PRIMARY KEY,
        newID text NOT NULL,
        curID text NOT NULL,
        curCrypt text NOT NULL,
        opStamp timestamp);
        """.format(self._tb_name)
        if not self.is_tb_exist(self._tb_name):
            self._db_cur.execute(tb_create_sql)

    def insert_rd(self, new_id, cur_id, cur_crypt):
        if not self.is_tb_exist(self._tb_name):
            self.create_tb(self._tb_name)
        insert_rd_sql = """
        INSERT INTO {} ('newID', 'curID', 'curCrypt', 'opStamp') 
        VALUES (?,?,?,?);
        """.format(self._tb_name)
        self._db_cur.execute(insert_rd_sql,
                             (new_id, cur_id, cur_crypt, datetime.now()))

    def checkout_rd(self, new_id):
        checkout_rd_sql = """
        SELECT curCrypt, opStamp FROM {} 
        WHERE newID=? ORDER BY opStamp DESC;
        """.format(self._tb_name)
        self._db_cur.execute(checkout_rd_sql, (new_id,))
        rows = self._db_cur.fetchall()
        return pd.DataFrame(rows, columns=["curCrypt", "opStamp"])

    def is_tb_exist(self, tb_name):
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
