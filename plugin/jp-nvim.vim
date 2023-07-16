if exists('g:loaded_jp_nvim')
  finish
endif
let g:loaded_jp_nvim = 1

let s:save_cpo = &cpo
set cpo&vim

let s:plugin_name = 'jp.nvim'
let s:plugin_root = fnamemodify(resolve(expand('<sfile>:p')), ':h:h')
let s:plugin_cmd = s:plugin_root . '/' . s:plugin_name

function! s:Require(host) abort
  return jobstart([s:plugin_cmd], {'rpc': v:true})
endfunction

call remote#host#Register(s:plugin_name, '', function('s:Require'))

" manifest
call remote#host#RegisterPlugin('jp.nvim', '0', [
\ {'type': 'command', 'name': 'JP', 'sync': 1, 'opts': {'nargs': '1'}},
\ ])

let &cpo = s:save_cpo
unlet s:save_cpo
