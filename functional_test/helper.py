from typing import *
import psycopg2


class DBHelper:
    def __init__(self, conn_str: str) -> None:
        self._conn = psycopg2.connect(conn_str)

    def drop_all_tables(self) -> None:
        print("drop all tables")
        cur = self._conn.cursor()
        cur.execute("DROP SCHEMA public CASCADE;")
        cur.execute("CREATE SCHEMA public;")
        cur.execute("GRANT ALL ON SCHEMA public TO postgres;")
        cur.execute("GRANT ALL ON SCHEMA public TO public;")
        self._conn.commit()
        cur.close()

    def close(self) -> None:
        self._conn.close()


