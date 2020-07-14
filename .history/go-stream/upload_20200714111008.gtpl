<html>
<head>
	<title>Upload file</title>
</head>
<body>
<form enctype="multipart/form-data" action="http://localhost:8000/upload" method="post">
	<input type="text" name="fileName" />
	<input type="file" name="uploadFile" />
	<input type="submit" value="upload" />
</form>
</body>
</html>