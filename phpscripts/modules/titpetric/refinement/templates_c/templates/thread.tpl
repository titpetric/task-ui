<?php $_v=&$this->vars;?>
<section class="refinement">
	
	<form class="form refinement-start" data-api="<?php echo $_v['base'];?>
/threads" data-method="POST" data-next="reload">
		<label for="refinement-want">What to design</label>
		<p class="faint" id="refinement-want-help">One paragraph, and the thread starts from it. Every comment after it is a Claude run over the artifacts as they stand.</p>
		<textarea id="refinement-want" name="want" rows="5" aria-describedby="refinement-want-help" placeholder="A database schema with strict separation of user, user_login_method and user_login_method_password." required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Start a thread</button>
		</div>
	</form>

	<div class="refinement-threads">
		<p class="faint">No threads yet.</p>
	</div>
</section>
