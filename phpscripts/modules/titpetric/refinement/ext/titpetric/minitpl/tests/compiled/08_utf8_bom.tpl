<?php $_v=&$this->vars;?>
This file contains utf8 BOM.

It also uses the <?php echo htmlspecialchars($_v['ldelim'], ENT_QUOTES);?>
load<?php echo htmlspecialchars($_v['rdelim'], ENT_QUOTES);?>
 construct:

<?php $this->push();$this->load("07_eval.tpl");$this->assign($_v);$this->render();$this->pop();?>
