def generic_error_string_builder(className: str, error_text: str, code: int = None) -> str:
    if code is None:
        return f"{className}\n{error_text}"
    else:
        return f"{className} ({code}) \n{error_text}"

class SQLException(Exception):
    def __init__(self, message, error_text=None, code=None):
        super(SQLException, self).__init__(message)
        self.code = code
        self.error_text = error_text

    def __str__(self):
        return generic_error_string_builder("SQLException", self.error_text, self.code)
    
class FileSystemException(Exception):
    def __init__(self, message, error_text=None, code=None):
        super(FileSystemException, self).__init__(message)
        self.code = code
        self.error_text = error_text
    
    def __str__(self):
        return generic_error_string_builder("FileSystemException", self.error_text, self.code)
