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
          m('div').addClass('Tags small mt-auto'),
          m('div').addClass('input-group').hide().append([
            m('input').addClass('TagsInput form-control'),
            m('button').text('ok').addClass('OK btn btn-outline-secondary').attr({type:'button'}),
          ]),
          m('div').addClass('IconButtons  mt-auto ms-auto').append([
            m('i').addClass('bi bi-tag').attr({title:'edit tags'}),
            m('i').addClass('bi bi-trash').attr({title:'delete'}),
            m('i').addClass('bi bi-download').attr({title:'download'}),
          ]),
        ]),
      ]),
    ]),
  ]);

  // 有些事件要在该组件被实体化之后添加才有效。
  self.init = () => {
    const tagsArea = $(self.id + ' .Tags');

    const tagGroup = addPrefix(file.Tags);
    const groupItem = cc('a');
    const groupLink = '/light/search?tags=' + encodeURIComponent(tagGroup);
    tagsArea.append(
      m(groupItem).text('tags:').attr({href:groupLink, target:'_blank'})
        .addClass('Tag link-secondary')
    );

    file.Tags.forEach(name => {
      const tagItem = cc('a');
      const tagLink = '/light/search?tags=' + encodeURIComponent(name);
      tagsArea.append(
        m(tagItem).text('#'+name).attr({href:tagLink, target:'_blank'})
          .addClass('Tag link-secondary')
      );
    });

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
    item.init();
  });  
};

FileList.clear = () => {
  $(FileList.id).html('');
};
