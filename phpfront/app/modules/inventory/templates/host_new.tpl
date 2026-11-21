{eval $data_head = $data.head}
{include components/page-head.tpl}
<section class="block">
<form class="form" method="post" action="{data.action}">
	<div class="field">
		<label for="host-address">Address</label>
		<input type="text" id="host-address" name="address" required spellcheck="false" placeholder="dev1, 10.0.0.2, or root@dev1">
		<small class="faint">The name or IP ssh signs in to. A user is optional: task-ui tries debian and root.</small>
	</div>
	<div class="field">
		<label for="host-name">Name</label>
		<input type="text" id="host-name" name="name" spellcheck="false" placeholder="the address">
		<small class="faint">What the pages call it. Empty is the address.</small>
	</div>
	<div class="form-foot"><span></span><span class="actions"><button class="btn btn-primary" type="submit"><i class="ic ic-plus" aria-hidden="true"></i>Add the host</button></span></div>
</form>
</section>
