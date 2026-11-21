<?php $_v=&$this->vars; if($_v['data_form']['err']){?>
<div class="alert bad" role="alert"><?php echo htmlspecialchars($_v['data_form']['err'], ENT_QUOTES);?>
</div><?php }else{?>
<form class="note-form run-form" data-run-form data-note="<?php echo htmlspecialchars($_v['data_form']['note'], ENT_QUOTES);?>
" data-block="<?php echo htmlspecialchars($_v['data_form']['index'], ENT_QUOTES);?>
"<?php if($_v['data_form']['folder']){?>
 data-folder="<?php echo htmlspecialchars($_v['data_form']['folder'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_form']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_form']['dir'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_form']['job']){?>
 data-job="<?php echo htmlspecialchars($_v['data_form']['job'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_form']['prompt']){?>
 data-prompt="<?php echo htmlspecialchars($_v['data_form']['prompt'], ENT_QUOTES);?>
"<?php }?>>
<?php if($_v['data_form']['title']){?>
<p class="form-title"><?php echo htmlspecialchars($_v['data_form']['title'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_form']['subtitle']){?>
<p class="form-subtitle faint"><?php echo htmlspecialchars($_v['data_form']['subtitle'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_form']['context']){?>
<div class="form-context prose"><?php echo htmlspecialchars($_v['data_form']['context'], ENT_QUOTES);?>
</div>
<?php }?><?php if($_v['data_form']['prerequisites']){?>
<ul class="form-prerequisites faint"><?php if(!empty($_v['data_form']['prerequisites']))foreach($_v['data_form']['prerequisites'] as $_v['item']){?><li><?php echo htmlspecialchars($_v['item'], ENT_QUOTES);?>
</li><?php }?></ul>
<?php }?><?php if(!empty($_v['data_form']['sections']))foreach($_v['data_form']['sections'] as $_v['section']){?><?php if($_v['section']['title']){?>
<h3 class="form-section"><?php echo htmlspecialchars($_v['section']['title'], ENT_QUOTES);?>
</h3>
<?php }?><?php if($_v['section']['description']){?>
<p class="faint form-section-note"><?php echo htmlspecialchars($_v['section']['description'], ENT_QUOTES);?>
</p>
<?php }?><?php if(!empty($_v['section']['fields']))foreach($_v['section']['fields'] as $_v['field']){?><?php if($_v['field']['component'] == "hidden"){?>
<input type="hidden" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['field']['value'], ENT_QUOTES);?>
" data-key="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" data-component="hidden"<?php if($_v['field']['when']){?>
 data-when="<?php echo htmlspecialchars($_v['field']['when'], ENT_QUOTES);?>
"<?php }?><?php if($_v['field']['required']){?>
 data-required<?php }?>>
<?php }else{?>
<div class="field" data-key="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" data-component="<?php echo htmlspecialchars($_v['field']['component'], ENT_QUOTES);?>
"<?php if($_v['field']['kind']){?>
 data-kind="<?php echo htmlspecialchars($_v['field']['kind'], ENT_QUOTES);?>
"<?php }?><?php if($_v['field']['when']){?>
 data-when="<?php echo htmlspecialchars($_v['field']['when'], ENT_QUOTES);?>
"<?php }?><?php if($_v['field']['required']){?>
 data-required<?php }?><?php if($_v['field']['hidden']){?>
 hidden<?php }?>>
<?php if($_v['field']['component'] == "bool"){?>
<label class="check"><input type="checkbox" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" data-type="toggle"<?php if($_v['field']['checked']){?>
 checked<?php }?>> <?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</label>
<?php }elseif($_v['field']['component'] == "multi"){?>
<span class="label"><?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</span>
<div class="checks"><?php if(!empty($_v['field']['options']))foreach($_v['field']['options'] as $_v['option']){?><label class="check"><input type="checkbox" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
" data-type="set"<?php if($_v['option']['selected']){?>
 checked<?php }?><?php if($_v['option']['next']){?>
 data-next="<?php echo htmlspecialchars($_v['option']['next'], ENT_QUOTES);?>
"<?php }?>> <span><?php echo htmlspecialchars($_v['option']['label'], ENT_QUOTES);?>
</span><?php if($_v['option']['recommended']){?>
 <span class="badge signal">recommended</span><?php }?><?php if($_v['option']['rationale']){?>
<small class="faint"><?php echo htmlspecialchars($_v['option']['rationale'], ENT_QUOTES);?>
</small><?php }?></label><?php }?></div>
<?php }elseif($_v['field']['component'] == "checklist"){?>
<span class="label"><?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</span>
<div class="checks"><?php if(!empty($_v['field']['options']))foreach($_v['field']['options'] as $_v['option']){?><label class="check"><input type="checkbox" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
" data-type="set"<?php if($_v['option']['selected']){?>
 checked<?php }?><?php if($_v['option']['next']){?>
 data-next="<?php echo htmlspecialchars($_v['option']['next'], ENT_QUOTES);?>
"<?php }?>> <span><?php echo htmlspecialchars($_v['option']['label'], ENT_QUOTES);?>
</span><?php if($_v['option']['recommended']){?>
 <span class="badge signal">recommended</span><?php }?><?php if($_v['option']['rationale']){?>
<small class="faint"><?php echo htmlspecialchars($_v['option']['rationale'], ENT_QUOTES);?>
</small><?php }?></label><?php }?></div>
<?php }elseif($_v['field']['kind'] == "qcm_free_text"){?>
<span class="label"><?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</span>
<div class="checks"><?php if(!empty($_v['field']['options']))foreach($_v['field']['options'] as $_v['option']){?><label class="check"><input type="radio" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
"<?php if($_v['option']['selected']){?>
 checked<?php }?><?php if($_v['option']['next']){?>
 data-next="<?php echo htmlspecialchars($_v['option']['next'], ENT_QUOTES);?>
"<?php }?>> <span><?php echo htmlspecialchars($_v['option']['label'], ENT_QUOTES);?>
</span><?php if($_v['option']['recommended']){?>
 <span class="badge signal">recommended</span><?php }?><?php if($_v['option']['rationale']){?>
<small class="faint"><?php echo htmlspecialchars($_v['option']['rationale'], ENT_QUOTES);?>
</small><?php }?></label><?php }?></div>
<?php }elseif($_v['field']['component'] == "textarea"){?>
<span class="label"><?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</span>
<textarea name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" rows="4" spellcheck="false" aria-label="<?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
"<?php if($_v['field']['placeholder']){?>
 placeholder="<?php echo htmlspecialchars($_v['field']['placeholder'], ENT_QUOTES);?>
"<?php }?><?php if($_v['field']['required']){?>
 required<?php }?>><?php echo htmlspecialchars($_v['field']['value'], ENT_QUOTES);?>
</textarea>
<?php }elseif($_v['field']['component'] == "number"){?>
<span class="label"><?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</span>
<input type="number" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['field']['value'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
"<?php if($_v['field']['placeholder']){?>
 placeholder="<?php echo htmlspecialchars($_v['field']['placeholder'], ENT_QUOTES);?>
"<?php }?><?php if($_v['field']['required']){?>
 required<?php }?>>
<?php }elseif($_v['field']['component'] == "text"){?>
<span class="label"><?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</span>
<input type="text" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['field']['value'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
"<?php if($_v['field']['placeholder']){?>
 placeholder="<?php echo htmlspecialchars($_v['field']['placeholder'], ENT_QUOTES);?>
"<?php }?><?php if($_v['field']['required']){?>
 required<?php }?>>
<?php }else{?>
<span class="label"><?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
</span>
<select name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['field']['label'], ENT_QUOTES);?>
"<?php if($_v['field']['required']){?>
 required<?php }?>><?php if(!$_v['field']['required']){?>
<option value=""></option><?php }?><?php if(!empty($_v['field']['options']))foreach($_v['field']['options'] as $_v['option']){?><option value="<?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
"<?php if($_v['option']['next']){?>
 data-next="<?php echo htmlspecialchars($_v['option']['next'], ENT_QUOTES);?>
"<?php }?><?php if($_v['option']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['option']['label'], ENT_QUOTES);?>
</option><?php }?></select>
<?php }?><?php if($_v['field']['free_text']){?>
<input class="free-text" type="text" name="<?php echo htmlspecialchars($_v['field']['key'], ENT_QUOTES);?>
.text" data-free-text maxlength="<?php echo htmlspecialchars($_v['field']['max_free_text_bytes'], ENT_QUOTES);?>
" placeholder="<?php echo htmlspecialchars($_v['field']['free_text_placeholder'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['field']['free_text_label'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['field']['help']){?>
<small class="faint"><?php echo htmlspecialchars($_v['field']['help'], ENT_QUOTES);?>
</small>
<?php }?></div>
<?php }?><?php }?><?php }?><div class="form-note"><textarea name="general_note" rows="2" placeholder="Anything else? (optional)" aria-label="A note with the answers"></textarea></div>
<div class="form-foot"><?php if($_v['data_form']['job']){?>
<span class="faint form-runs">Runs <?php if($_v['data_form']['folder']){ echo htmlspecialchars($_v['data_form']['folder'], ENT_QUOTES);?>
/<?php }?><?php echo htmlspecialchars($_v['data_form']['job'], ENT_QUOTES);?>
</span><?php }?><span class="faint form-state"></span><button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_form']['label'], ENT_QUOTES);?>
</button></div>
</form><?php }?>
