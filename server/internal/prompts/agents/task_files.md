{{if eq .Language "日本語"}}（Persona：レポート、文書、表、画像、アーカイブなどをユーザーに渡す場合は、内容を返信に貼り付けず、最後の返信に次の形式で別の行に書いてください。
file: <このコンピューター上の絶対パス>
1 ファイルにつき 1 行、最大 5 件です。通常のファイルのみ指定し、ディレクトリは先にアーカイブしてください。Persona が会話にファイルを届け、ユーザーはスマートフォンや Web で開けます。長いファイルやテキスト以外のファイルは、返信に貼ると途中で切れるため貼り付けないでください。）{{else if .Language}}(Persona: To hand the user a file such as a report, document, spreadsheet, image, or archive, put its path on a separate line in your final reply instead of pasting its contents.
file: <absolute path on this computer>
Use one line per file, up to five regular files. Archive directories first. Persona delivers the files to the conversation so the user can open them on their phone or the web. Pasted contents can be cut off, especially for long or binary files.){{else}}（Persona：任务的结果是要交给用户的文件时——报告、文档、表格、图片、压缩包等——不要让用户看内容贴出来的版本：在最后一轮回复里单独写一行
file: <这台电脑上的绝对路径>
有几个写几行，最多 5 个；只写普通文件（目录先打包）。Persona 会把它们送到对话那边，用户的手机和网页上都能打开。文件长或者不是文本时更不要贴：贴进回复的内容会被截断，也传不过去。）{{end}}
