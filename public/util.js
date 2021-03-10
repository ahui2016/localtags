"use strict"

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
      const msg = `${this.status} ${this.responseText}`;
      if (options.alerts) {
        options.alerts.Insert('danger', msg);
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

function CreateAlerts(max) {
  if (!max) max = 5;
  const alerts = cc('div');

  alerts.count = 0;

  alerts.insertElem = (elem) => {
    $(alerts.id).prepend(elem);
    alerts.count++;
    if (alerts.count > max) {
      $(`${alerts.id} p:last-of-type`).remove();
    }
  };

  alerts.insert = (msgType, msg) => {
    const elem = m('p').addClass(`alert alert-${msgType}`).append([
      m('span').text(dayjs().format('HH:mm:ss')),
      m('span').text(msg),
    ]);
    alerts.insertElem(elem);
  };

  alerts.clear = () => {
    $(alerts.id).html('');
    alerts.count = 0;
  };

  alerts.view = () => m('div').attr({id: alerts.raw_id}).addClass('alerts');
  return alerts;
}
