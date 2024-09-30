class SQLException(Exception):
    def __init__(self, message, code=None):
        super(SQLException, self).__init__(message)
        self.code = code