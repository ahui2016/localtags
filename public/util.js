"use strict"

const tempFolder = "/temp/"
const thumbSuffix  = ".small.jpg"
const thumbsFolder = "/thumbs/"

function getTempThumb(id) {
  return tempFolder + id + thumbSuffix;
}

function getThumbURL(id) {
  return thumbsFolder + id;
}

// make a new vnode by name, or return its view.
function m(name) {
  if (jQuery.type(name) == 'string') {
    return $(document.createElement(name));
  }
  return name.view();
}

// cc creates a component with an id.
function cc(name, id, elements) {
  if (!id) id = '' + Math.round(Math.random() * 100000000);
  const vnode = m(name).attr('id', id);
  if (elements) vnode.append(elements);
  return {id: '#'+id, raw_id: id, view: () => vnode};
}

function disable(id) { $(id).prop('disabled', true); }

function enable(id) { $(id).prop('disabled', false); }

// options = { method, url, body, alerts, buttonID }
function ajax(options, onSuccess, onFail, onAlways) {
  if (options.buttonID) disable(options.buttonID);
  const xhr = new XMLHttpRequest();
  xhr.open(options.method, options.url);
  xhr.onerror = () => {
    window.alert('An error occurred during the transaction');
  };
  xhr.addEventListener('load', function() {
    if (this.status == 200) {
      if (onSuccess) {
        const resp = this.responseText ? JSON.parse(this.responseText) : null;
        onSuccess(resp);
      }
    } else {
      let msg;
      if (this.responseText) {
        const resp = JSON.parse(this.responseText);
        msg = resp.message ? resp.message : `${this.status} ${this.responseText}`
      } else {
        msg = `${this.status} ${this.statusText}`
      }
      if (options.alerts) {
        options.alerts.insert('danger', msg);
      } else {
        console.log(msg);
      }
      if (onFail) onFail(this);
    }
  });
  xhr.addEventListener('loadend', function() {
    if (options.buttonID) enable(options.buttonID);
    if (onAlways) onAlways(this);
  });
  xhr.send(options.body);
}

// 获取地址栏的参数。
function getUrlParam(param) {
  let loc = new URL(document.location);
  return loc.searchParams.get(param);
}

// 把文件大小换算为 KB 或 MB
function fileSizeToString(fileSize, fixed) {
  if (fixed == null) {
    fixed = 2
  }
  const sizeMB = fileSize / 1024 / 1024;
  if (sizeMB < 1) {
    return `${(sizeMB * 1024).toFixed(fixed)} KB`;
  }
  return `${sizeMB.toFixed(fixed)} MB`;
}

function addPrefix(setOrArr, prefix) {
  if (!setOrArr) return '';
  let arr = Array.from(setOrArr);
  if (!prefix) prefix = '';
  return arr.map(x => prefix + x).join(' ');
}

function tag_replace(tags) {
  return tags.replace(/[#;,，'"/\+\n]/g, ' ').trim();
}

function tagsStringToSet(tags) {
  const trimmed = tag_replace(tags);
  if (trimmed.length == 0) return new Set();
  const arr = trimmed.split(/ +/);
  return new Set(arr);
}

function tagsStringToArray(tags) {
  return Array.from(tagsStringToSet(tags));
}

function getThumbByFiletype(filetype) {
  let prefix = filetype.split('/').shift();
  let suffix = filetype.split('/').pop();
  switch (suffix) {
    case 'doc':
    case 'docx':
      return '/public/icons/file-earmark-word.jpg';
    case 'xls':
    case 'xlsx':
      return '/public/icons/file-earmark-excel.jpg';
    case 'ppt':
    case 'pptx':
      return '/public/icons/file-earmark-ppt.jpg';
    default:
      switch (prefix) {
        case 'image':
          return '/public/icons/file-earmark-image.jpg';
        case 'video':
          return '/public/icons/file-earmark-play.jpg';
        case 'office':
        case 'ebook':
          return '/public/icons/file-earmark-richtext.jpg';
        case 'compressed':
          return '/public/icons/file-earmark-zip.jpg';
        case 'text':
          return '/public/icons/file-earmark-text.jpg';
        case 'audio':
          return '/public/icons/file-earmark-music.jpg';
        default:
          return '/public/icons/file-earmark-binary.jpg';
      }    
  }
}


/* compoents */

const Spacer = { view: () => $('<div style="margin-bottom: 2em;"></div>') };

const BottomLine = { view: () => $('<div style="margin-top: 200px;"></div>') };

const Loading = {
  view: () => $('<p id="loading" class="alert-info">Loading...</p>'),
  hide: () => { $('#loading').hide(); },
  reset: (text) => {
    if (!text) {
      $('#loading').show();
      return;
    }
    $('#loading').show().text(text); 
  },
};

function CreateInfoPair(name, msg) {
  const infoMsg = {
    id: `#about-${name}-msg`,
    view: () => $(`<div id="about-${name}-msg" class="InfoMessage" style="display:none">${msg}</div>`),
    toggle: () => { $(infoMsg.id).toggle(); },
    setMsg: (msg) => { $(infoMsg.id).text(msg); },
  };
  const infoIcon = {
    id: `#about-${name}-icon`,
    view: () => $(`<img id= "about-${name}-icon" src="/public/info-circle.svg" class="IconButton" alt="info" title="显示/隐藏说明">`)
    .click(infoMsg.toggle),
  };
  return [infoIcon, infoMsg];
}

function CreateAlerts() {
  const alerts = cc('div');

  alerts.insertElem = (elem) => {
    $(alerts.id).prepend(elem);
  };

  alerts.insert = (msgType, msg) => {
    const time = dayjs().format('HH:mm:ss');
    const elem = m('div')
      .addClass(`alert alert-${msgType} alert-dismissible fade show`)
      .attr({role:'alert'})
      .append([
        m('span').text(`${time} ${msg}`),
        m('button').attr({type: 'button', class: "btn-close", 'data-bs-dismiss': "alert", 'aria-label':"Close"}),
      ]);
    alerts.insertElem(elem);
  };

  alerts.clear = () => {
    $(alerts.id).html('');
  };

  return alerts;
}
