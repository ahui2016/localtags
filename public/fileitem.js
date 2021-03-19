function FileItem(file) {
  let thumbClass;
  if (file.Thumb) {
    thumbClass = 'col-md-2 col-3';
    file.Thumb = getThumbURL(file.ID);
  } else {
    thumbClass = 'col-md-2 col-2';
    file.Thumb = getThumbByFiletype(file.Type);
  }

  let cardSubtitle;
  if (file.ID.length > 10) {
    cardSubtitle = `(size: ${fileSizeToString(file.Size)})`;
  } else {
    cardSubtitle = `id:${file.ID} (size: ${fileSizeToString(file.Size)})`;
  }
  
  const self = cc('div', file.ID);

  self.view = () => m('div').attr({id: self.raw_id}).addClass('FileItem card mb-3').append([
    m('div').addClass('row g-0').append([
      m('div').addClass(thumbClass).append([
        m('img').addClass('card-img ').attr({src: file.Thumb}),
      ]),
      m('div').addClass('col').append([
        m('div').addClass('card-body d-flex flex-column h-100').append([
          m('p').addClass('small card-subtitle text-muted').text(cardSubtitle),
          m('p').addClass('card-text mb-0').text(file.Name),
          m('div').addClass('Tags small text-muted mt-auto').text(addPrefix(file.Tags, '#')),
          m('div').addClass('input-group mt-auto').hide().append([
            m('input').addClass('TagsInput form-control'),
            m('button').text('ok').addClass('OK btn btn-outline-secondary').attr({type:'button'}),
          ]),
        ]),
      ]),
    ]),
  ]);

  // 有些事件要在该组件被实体化之后添加才有效。
  self.attachEvents = () => {
    const tagsText = $(self.id + ' .Tags');
    const tagsInput = $(self.id + ' .TagsInput');
    const inputGroup = $(self.id + ' .input-group');

    $(self.id + ' .OK').click(() => {
      const tags = tagsInput.val();
      const tagsSet = tagsStringToSet(tags);
      tagsText.show().text(addPrefix(tagsSet, '#'));
      inputGroup.hide();
    });
    
    tagsText.dblclick(() => {
      inputGroup.show();
      tagsInput.val(tagsText.text()).focus();
      tagsText.hide();
    });  
  };
  
  return self;
}

const FileList = cc('div');

FileList.prepend = (files) => {
  files.forEach(file => {
    const item = FileItem(file);
    $(FileList.id).prepend(m(item));
    item.attachEvents();
  });  
};

FileList.clear = () => {
  $(FileList.id).html('');
};
