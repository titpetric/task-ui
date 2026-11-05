			<tr id="key-{data_key.name}">
				<td class="key">{if $data_key.href}<a class="mono" href="{data_key.href}">{data_key.name}</a>{else}<span class="mono">{data_key.name}</span>{/if}{if $data_key.encrypts} <span class="badge signal">encryption</span>{/if}{if $data_key.backup}<span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: {data_key.backup}</span>{/if}{if $data_key.comment}<div class="faint">{data_key.comment}</div>{/if}</td>
				<td class="mono hide-narrow" data-label="Algorithm">{data_key.algo}</td>
				<td class="mono nowrap hide-narrow" data-label="Size">{data_key.size} bits</td>
				<td class="pubkey" data-label="Public key">{eval $data_copy = $data_key.copy}{include components/copyable.tpl}</td>
				<td class="actions">{eval $data_menu = $data_key.menu}{include components/menu.tpl}</td>
			</tr>
