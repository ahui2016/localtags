function FileItem(file) {
  let thumbClass;
  if (file.Thumb) {
    thumbClass = 'col-md-2 col-2';
    file.Thumb = getThumbURL(file.ID);
  } else {
    thumbClass = 'col-md-2 col-1';
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
              m('i').addClass('bi bi-bootstrap-reboot').attr({title:'restore'}).hide(),
              m('i').addClass('bi bi-trash-fill').attr({title:'permanently delete'}).hide(),
              m('i').addClass('bi bi-download').attr({title:'download'}),
            ]),
            m('div').text('RESTORED').addClass('Restored mt-auto ms-auto').hide(),
            m('div').text('DELETED').addClass('Deleted mt-auto ms-auto').hide(),
          ]),
        ]),
      ]),
    ]),
    m(ItemAlerts),
  ]);


  const thumb_id = self.id + ' .card-img';
  const filename_id = self.id + ' .Filename';
  const filename = $(filename_id);
  const name_input_id = self.id + ' .NameInputGroup';
  const tags_id = self.id + ' .Tag';
  const buttons_id = self.id + ' .IconButtons';

  const tags_btn_id = self.id + ' .bi-tag';
  const del_btn_id = self.id + ' .bi-trash';
  const restore_btn_id = self.id + ' .bi-bootstrap-reboot';
  const really_del_btn_id = self.id + ' .bi-trash-fill';
  const dl_btn_id = self.id + ' .bi-download';

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
      const tagLink = '/light/tag?name=' + encodeURIComponent(name);
      tagsArea.append(
        m(tagItem).text('#'+name).attr({href:tagLink, target:'_blank'})
          .addClass('Tag link-secondary')
      );
    });
  }

  self.toggleTagsArea = () => {
    $(self.id + ' .Tags').toggle();
    $(self.id + ' .TagsInputGroup').toggle();
    $(buttons_id).toggle();
  }

  self.toggleFilename = () => {
    $(filename_id).toggle();
    $(name_input_id).toggle();
  }

  self.afterDeleted = () => {
    $(thumb_id).css('filter', 'opacity(0.5) grayscale(1)');
    $(name_input_id).hide();
    $(filename_id).show().addClass('text-secondary')
    disable(filename_id);
    disable(tags_id);
    $(buttons_id).hide();
    $(self.id + ' .Deleted').show();
  };

  // 有些事件要在该组件被实体化之后添加才有效。
  self.init = () => {
    const tagsInput = $(self.id + ' .TagsInput');
    const nameInput = $(self.id + ' .NameInput');

    if (file.Deleted) {
      disable(filename_id);
      $(tags_btn_id).hide();
      $(del_btn_id).hide();
      $(restore_btn_id).show();
      $(really_del_btn_id).show();
    }
    
    self.resetTags(file.Tags);
    
    $(tags_btn_id).click(() => {
      self.toggleTagsArea();
      tagsInput.val(addPrefix(self.tags, '#')).focus();
    });

    const tags_ok_id = self.id+' .TagsOK';
    $(tags_ok_id).click(() => {
      const tagsSet = tagsStringToSet(tagsInput.val());
      if (tagsSet.size == 0 || eqSets(tagsSet, self.tags)) {
        self.toggleTagsArea();
        return;
      }
      const body = new FormData();
      body.append('id', file.ID);
      body.append('tags', JSON.stringify(Array.from(tagsSet)));
      ajax({method:'POST',url:'/api/update-tags',alerts:ItemAlerts,buttonID:tags_ok_id,body:body},
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

    const body = new FormData();
    body.append('id', file.ID);

    $(restore_btn_id).click(() => {
      ajax({method:'POST',url:'/api/undelete-file',alerts:ItemAlerts,buttonID:restore_btn_id,body:body},
          () => {
            $(buttons_id).hide();
            $(self.id + ' .Restored').show();
          });
    });

    $(del_btn_id).click(() => {
      ajax({method:'POST',url:'/api/delete-file',alerts:ItemAlerts,buttonID:del_btn_id,body:body},
          () => {
            self.afterDeleted();
          });
    });

    $(really_del_btn_id).click(() => {
      ItemAlerts.insert('info', '再点击一次删除按钮彻底删除该文件，不可恢复。');
      $(really_del_btn_id).off();
      window.setTimeout(() => {
        $(really_del_btn_id).click(() => {
          ajax({method:'POST',url:'/api/really-delete-file',alerts:ItemAlerts,buttonID:really_del_btn_id,body:body},
          () => {
            self.afterDeleted();
          });
        });
      }, 1000);
    });

    filename.dblclick(() => {
      self.toggleFilename();
      nameInput.val(filename.text()).focus();
    });
    const name_ok_id = self.id+' .NameOK';
    $(name_ok_id).click(() => {
      const oldName = filename.text();
      const newName = nameInput.val();
      if (newName.length == 0 || newName == oldName) {
        self.toggleFilename();
        return;
      }
      const body = new FormData();
      body.append('id', file.ID);
      body.append('name', newName);
      ajax({method:'POST',url:'/api/rename-file',alerts:ItemAlerts,buttonID:name_ok_id,body:body},
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
