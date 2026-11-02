<?php $_v=&$this->vars;?>
We check if arrays are empty to suppress warnings/notices.

<?php foreach(array("GET", "POST") as $_v['method']){?>
HTTP <?php echo htmlspecialchars($_v['method'], ENT_QUOTES);}foreach(array("GET", "POST") as $_v['method']){?>
HTTP <?php echo htmlspecialchars($_v['method'], ENT_QUOTES);}if(!empty($_v['methods']))foreach($_v['methods'] as $_v['method']){?>
HTTP <?php echo htmlspecialchars($_v['method'], ENT_QUOTES);}?>
