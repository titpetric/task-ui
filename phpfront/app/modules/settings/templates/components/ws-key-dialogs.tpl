{* input scope: data_wsdialogs. See ws-key-dialogs.md. *}
<dialog class="confirm" id="key-generate-dialog" aria-labelledby="key-generate-title">
	{eval $data_wsgen = $data_wsdialogs.generate}{include components/ws-key-generate.tpl}</dialog>
<dialog class="confirm" id="key-upload-dialog" aria-labelledby="key-upload-title">
	{eval $data_wsup = $data_wsdialogs.upload}{include components/ws-key-upload.tpl}</dialog>
