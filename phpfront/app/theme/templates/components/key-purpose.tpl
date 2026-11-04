{* input scope: none. See key-purpose.md. *}
<div class="field" role="radiogroup" aria-labelledby="key-purpose-label">
	<span class="label" id="key-purpose-label">Purpose</span>
	<label class="check"><input type="radio" name="purpose" value="ssh" checked> SSH key</label>
	<small class="faint">Runs, git and logins authenticate with it.</small>
	<label class="check"><input type="radio" name="purpose" value="encrypt"> Encryption key</label>
	<small class="faint">ed25519 or RSA; encrypts the secrets, never handed out.</small>
</div>