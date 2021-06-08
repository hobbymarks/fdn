import datetime
from pathlib import Path

# From Third Party
import pandas as pd
from sqlalchemy import Column, DateTime, Integer, String, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()


class _FDNRecord(Base):
    __tablename__ = "fdnrd"
    id = Column(Integer, primary_key=True)
    newID = Column(String, nullable=False)
    curID = Column(String, nullable=False)
    curCrypt = Column(String, nullable=False)
    sepDirPathID = Column(String)
    absDirPathID = Column(String)
    opStamp = Column(DateTime,
                     nullable=False,
                     default=datetime.datetime.utcnow())


class FDNDB:
    def __init__(self, db_path: Path):
        self.engine = create_engine("sqlite:///" + str(db_path))
        Base.metadata.create_all(self.engine)
        DBSession = sessionmaker(bind=self.engine)
        self.session = DBSession()

    def insert_rd(self, new_id: str, cur_id: str, cur_crypt: str,
                  sep_dp_id: str, abs_dp_id: str):
        self.session.add(
            _FDNRecord(newID=new_id,
                       curID=cur_id,
                       curCrypt=cur_crypt,
                       sepDirPathID=sep_dp_id,
                       absDirPathID=abs_dp_id))
        self.session.commit()

    def checkout_rd(self, new_id: str):
        rds = self.session.query(_FDNRecord).filter(
            _FDNRecord.newID == new_id).all()
        rows = [[rd.curCrypt, rd.opStamp] for rd in rds]
        return pd.DataFrame(rows, columns=["curCrypt", "opStamp"])
