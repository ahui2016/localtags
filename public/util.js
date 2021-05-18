"use strict"

const tempFolder = "/temp/"
const thumbSuffix  = ".small.jpg"
const thumbsFolder = "/thumbs/"
const mainBucket = "/mainbucket/"

function getTempThumb(id) {
  return tempFolder + id + thumbSuffix;
}

function getThumbURL(id) {
  return thumbsFolder + id;
}

function getPreviewURL(id, type) {
  if (type == 'text/md') {
    return '/light/md-preview?id='+id;
  }
  return mainBucket + id;
}

function getPhotoURL(id) {
  return mainBucket + id;
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
  if (!id) id = 'r' + Math.round(Math.random() * 100000000);
  const vnode = m(name).attr('id', id);
  if (elements) vnode.append(elements);
  return {id: '#'+id, raw_id: id, view: () => vnode};
}

function hide(id) {
  $(id).addClass('d-none');
}

function show(id) {
  $(id).removeClass('d-none');
}

function toggle(id) {
  $(id).toggleClass('d-none');
}

function disable(id) {
  const nodeName = $(id).prop('nodeName');
  if (nodeName == 'BUTTON' || nodeName == 'INPUT') {
    $(id).prop('disabled', true); 
  } else {
    $(id).css('pointer-events', 'none');
  }
}

function enable(id) {
  const nodeName = $(id).prop('nodeName');
  if (nodeName == 'BUTTON' || nodeName == 'INPUT') {
    $(id).prop('disabled', false);
  } else {
    $(id).css('pointer-events', 'auto');
  }
}

// options = { method, url, body, alerts, buttonID, responseType }
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
        if (options.responseType && options.responseType == 'text') {
          onSuccess(this.responseText);
          return;
        }
        const resp = this.responseText ? JSON.parse(this.responseText) : null;
        onSuccess(resp);
      }
    } else {
      let msg;
      try {
        const resp = JSON.parse(this.responseText);
        msg = resp.message ? resp.message : `${this.status} ${this.responseText}`;
      } catch {
        msg = `${this.status} ${this.responseText}`;
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

function scrollTop(id) {
  $('html, body').scrollTop($(id).offset().top - 40);
}

// 获取地址栏的参数。
function getUrlParam(param) {
  let loc = new URL(document.location);
  return loc.searchParams.get(param);
}

// 把文件大小转换为方便人类阅读的格式。
function fileSizeToString(fileSize, fixed) {
  if (fixed == null) {
    fixed = 2
  }
  const sizeGB = fileSize / 1024 / 1024 / 1024;
  if (sizeGB < 1) {
    const sizeMB = sizeGB * 1024;
    if (sizeMB < 1) {
      const sizeKB = sizeMB * 1024;
      return `${sizeKB.toFixed(fixed)} KB`;
    }
    return `${sizeMB.toFixed(fixed)} MB`;
  }
  return `${sizeGB.toFixed(fixed)} GB`;
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

// checks if two sets are equal
function eqSets(a, b) {
  if (a.size != b.size) return false;
  for (const item of a) {
    if (!b.has(item)) return false;
  }
  return true;
}

function isPreviewable(file) {
  const prefix = file.Type.split('/').shift();
  const suffix = file.Name.split('.').pop();
  if (prefix == 'image') return true;
  switch (prefix) {
    case 'image':
    case 'text':
      return true;
    default:
      switch (suffix) {
        case 'mp3':
        case 'mp4':
        case 'pdf':
          return true;
        default:
          return false;
      }
  }
}

function getThumbByFiletype(filetype) {
  let prefix = filetype.split('/').shift();
  let suffix = filetype.split('/').pop();
  switch (suffix) {
    case 'md':
      return '/public/icons/file-earmark-md.jpg';
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
  view: () => m('div').attr({id:'loading'}).addClass('text-center').append([
    m('div').addClass('spinner-border').attr({role:'status'}).append(
      m('span').addClass('visually-hidden').text('Loading...')
    ),
  ]),
  hide: () => { $('#loading').hide(); },
  show: () => { $('#loading').show(); },
};

function CreateInfoPair(name, messages) {
  
  const infoMsg = cc('div', 'abount'+name+'msg');
  infoMsg.view = () => m('div').attr({id:InfoMsg.raw_id}).addClass('card text-dark bg-light my-3').append([
    m('div').text(name).addClass('card-header'),
    m('div').addClass('card-body text-secondary').append(
      m('div').addClass('card-text').append(messages),
    ),
  ]);
  infoMsg.setMsg = (messages) => {
    $(infoMsg.id + ' .card-text').html('').append(messages);
  };
  const infoBtn = {
    view: () => m('i').addClass('bi bi-info-circle').css({cursor:'pointer'})
    .attr({title:'显示/隐藏'+name}).click(() => { $(infoMsg.id).toggle() }),
  }
  return [infoBtn, infoMsg];
}

function CreateAlerts() {
  const alerts = cc('div');

  alerts.insertElem = (elem) => {
    $(alerts.id).prepend(elem);
  };

  alerts.insert = (msgType, msg) => {
    const time = dayjs().format('HH:mm:ss');
    const time_and_msg = `${time} ${msg}`;
    if (msgType == 'danger') {
      console.log(time_and_msg);
    }
    const elem = m('div')
      .addClass(`alert alert-${msgType} alert-dismissible fade show mt-1 mb-0`)
      .attr({role:'alert'})
      .append([
        m('span').text(time_and_msg),
        m('button').attr({type: 'button', class: "btn-close", 'data-bs-dismiss': "alert", 'aria-label':"Close"}),
      ]);
    alerts.insertElem(elem);
  };

  alerts.clear = () => {
    $(alerts.id).html('');
  };

  return alerts;
}
