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

    def get_email_verification_code(self, email: str) -> str:
        cur = self._conn.cursor()
        cur.execute(
            "SELECT verification_code FROM email_verifications WHERE email = %s AND expired_at > now()", (email,))
        data = cur.fetchone()
        self._conn.commit()
        cur.close()
        return data[0]

    def close(self) -> None:
        self._conn.close()
