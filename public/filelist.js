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
  
  const ItemAlerts = CreateAlerts();
  const self = cc('div', 'f'+file.ID);

  self.view = () => m('div').attr({id: self.raw_id}).addClass('FileItem mb-3').append([
    m('div').addClass('card').append([
      m('div').addClass('row g-0').append([
        m('div').addClass(thumbClass).append([
          m('img').addClass('card-img').attr({src: file.Thumb}),
        ]),
        m('div').addClass('col').append([
          m('div').addClass('card-body d-flex flex-column h-100').append([
            m('p').addClass('small card-subtitle text-muted').text(cardSubtitle),
            m('p').addClass('Filename card-text text-break').text(file.Name),
            m('div').addClass('NameInputGroup input-group').hide().append([
              m('input').addClass('NameInput form-control'),
              m('button').text('ok').addClass('NameOK btn btn-outline-secondary').attr({type:'button'}),
            ]),
            m('div').addClass('Tags small'),
            m('div').addClass('TagsInputGroup input-group').hide().append([
              m('input').addClass('TagsInput form-control'),
              m('button').text('ok').addClass('TagsOK btn btn-outline-secondary').attr({type:'button'}),
            ]),
            m('div').addClass('IconButtons mt-auto ms-auto').append([
              m('i').addClass('bi bi-tag').attr({title:'edit tags'}),
              m('i').addClass('bi bi-trash').attr({title:'delete'}),
              m('i').addClass('bi bi-download').attr({title:'download'}),
            ]),
            m('div').addClass('Deleted mt-auto ms-auto').hide().append(
              m('span').text('DELETED').addClass('badge bg-secondary')
            ),
          ]),
        ]),
      ]),
    ]),
    m(ItemAlerts),
  ]);

  self.tags = new Set();

  self.resetTags = (tags) => {
    self.tags = new Set(tags);
    const tagGroup = addPrefix(self.tags);
    const groupItem = cc('a');
    const groupLink = '/light/search?tags=' + encodeURIComponent(tagGroup);
  
    const tagsArea = $(self.id + ' .Tags');
    tagsArea.html('');
    tagsArea.append(
      m(groupItem).text('tags:').attr({href:groupLink, target:'_blank'})
        .addClass('Tag link-secondary')
    );
  
    self.tags.forEach(name => {
      const tagItem = cc('a');
      const tagLink = '/light/search?tags=' + encodeURIComponent(name);
      tagsArea.append(
        m(tagItem).text('#'+name).attr({href:tagLink, target:'_blank'})
          .addClass('Tag link-secondary')
      );
    });
  }

  self.toggleTagsArea = () => {
    const tagsArea = $(self.id + ' .Tags');
    const tagsInputGroup = $(self.id + ' .TagsInputGroup');
    const buttons = $(self.id + ' .IconButtons');
    tagsArea.toggle();
    tagsInputGroup.toggle();
    buttons.toggle();
  }

  self.toggleFilename = () => {
    const filename = $(self.id + ' .Filename');
    const nameInputGroup = $(self.id + ' .NameInputGroup');
    filename.toggle();
    nameInputGroup.toggle();
  }

  // 有些事件要在该组件被实体化之后添加才有效。
  self.init = () => {
    const tagsInput = $(self.id + ' .TagsInput');
    const buttons = $(self.id + ' .IconButtons');
    const tagsBtn = $(self.id + ' .bi-tag');
    
    self.resetTags(file.Tags);
    
    tagsBtn.click(() => {
      self.toggleTagsArea();
      tagsInput.val(addPrefix(self.tags, '#')).focus();
    });

    $(self.id + ' .TagsOK').click(() => {
      const tags = tagsInput.val();
      const tagsSet = tagsStringToSet(tags);
      if (tagsSet.size == 0 || eqSets(tagsSet, self.tags)) {
        self.toggleTagsArea();
        return;
      }
      const body = new FormData();
      body.append('id', file.ID);
      body.append('tags', JSON.stringify(Array.from(tagsSet)));
      ajax({method:'POST',url:'/api/update-tags',alerts:ItemAlerts,buttonID:self.id+' .TagsOK',body:body},
          () => {
            // onsuccess
            self.toggleTagsArea();
            self.resetTags(tagsSet);
          },
          () => {
            // onfail
            tagsInput.focus();
          });
    });

    const deleteBtn = $(self.id + ' .bi-trash');
    const thumb = $(self.id + ' .card-img');
    const deleted = $(self.id + ' .Deleted');
    const filename = $(self.id + ' .Filename');
    const tags = $(self.id + ' .Tag');
    const body = new FormData();
    body.append('id', file.ID);
    deleteBtn.click(() => {
      buttons.hide();
      ajax({method:'POST',url:'/api/delete-file',alerts:ItemAlerts,body:body},
          () => {
            // onsuccess
            thumb.css('filter', 'opacity(0.5) grayscale(1)');
            filename.addClass('text-secondary');
            tags.removeAttr('href');
            deleted.show();
          },
          () => {
            // onfail
            buttons.show();
          });
    });

    const nameInput = $(self.id + ' .NameInput');

    filename.dblclick(() => {
      self.toggleFilename();
      nameInput.val(filename.text()).focus();
    });
    $(self.id + ' .NameOK').click(() => {
      const oldName = filename.text();
      const newName = nameInput.val();
      if (newName.length == 0 || newName == oldName) {
        self.toggleFilename();
        return;
      }
      const body = new FormData();
      body.append('id', file.ID);
      body.append('name', newName);
      ajax({method:'POST',url:'/api/rename-file',alerts:ItemAlerts,buttonID:self.id+' .NameOK',body:body},
          () => {
            // onsuccess
            self.toggleFilename();
            filename.text(newName);
          },
          () => {
            // onfail
            nameInput.focus();
          });
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
