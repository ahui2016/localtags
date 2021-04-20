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
    cardSubtitle = `id: ${file.ID} (size: ${fileSizeToString(file.Size)})`;
  }

  const ctime = dayjs.unix(file.CTime).format('MMMM D, YYYY HH:mm:ss');
  
  const ItemAlerts = CreateAlerts();
  const self = cc('div', 'f'+file.ID);

  self.view = () => m('div').attr({id: self.raw_id}).addClass('FileItem mb-3').append([
    m('div').addClass('card').append([
      m('div').addClass('row g-0').append([
        m('div').addClass(thumbClass).append([
          m('img').addClass('NoLinkImg card-img').attr({src: file.Thumb}),
          m('a').attr({href:getPhotoURL(file.ID),target:'_blank'}).hide().addClass('LinkImg').append(
            m('img').addClass('card-img').attr({src: file.Thumb}),
          ),
        ]),
        m('div').addClass('col').append([
          m('div').addClass('card-body d-flex flex-column h-100').append([
            m('p').addClass('small text-muted mb-0').append([
              m('span').text(cardSubtitle).attr({title:ctime}).css('cursor','crosshair'), ' ',
              m('span').text('DAMAGED').addClass('Damaged text-light bg-danger px-1').hide(),
            ]),
            m('p').addClass('FilenameArea card-text text-break').append([
              m('a').text(`[${file.Count}]`).addClass('FileCount link-secondary text-decoration-none')
                .attr({title:'same name files',href:'/light/search?fileid=' + file.ID}).hide(),
              m('span').addClass('Filename').text(file.Name),
            ]),
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
              m('i').addClass('bi bi-bootstrap-reboot').attr({title:'restore'}).hide(),
              m('i').addClass('DeleteBtn bi bi-trash').attr({title:'delete'}),
              m('i').text('DeleteRecycle').addClass('DeleteRecycle bi bi-trash').hide(),
              m('i').addClass('bi bi-download').attr({title:'download'}).click(self.download),
            ]),
            m('div').text('RESTORED').addClass('Restored mt-auto ms-auto').hide(),
            m('div').text('DELETED').addClass('Deleted mt-auto ms-auto').hide(),
          ]),
        ]),
      ]),
    ]),
    m(ItemAlerts),
  ]);

  const no_link_img_id = self.id + ' .NoLinkImg';
  const link_img_id = self.id + ' .LinkImg';
  const filename_id = self.id + ' .Filename';
  const filename_area_id = self.id + ' .FilenameArea';
  const name_input_id = self.id + ' .NameInputGroup';
  const file_count_id = self.id + ' .FileCount';
  const tags_id = self.id + ' .Tag';
  const buttons_id = self.id + ' .IconButtons';

  const tags_btn_id = self.id + ' .bi-tag';
  const del_btn_id = self.id + ' .DeleteBtn';
  const restore_btn_id = self.id + ' .bi-bootstrap-reboot';
  const del_recycle_id = self.id + ' .DeleteRecycle';
  const dl_btn_id = self.id + ' .bi-download';

  self.download = () => {
    ajax({method:'GET',url:'/api/download/'+file.ID,alerts:ItemAlerts,buttonID:dl_btn_id},
        (resp) => {
          ItemAlerts.insert('success', resp.message);
        });
  };

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
    $(filename_area_id).toggle();
    $(name_input_id).toggle();
  }

  self.afterDeleted = () => {
    $(self.id + ' .card-img').css('filter', 'opacity(0.5) grayscale(1)');
    $(name_input_id).hide();
    $(filename_area_id).show();
    $(filename_id).addClass('text-secondary');
    disable(link_img_id);
    disable(file_count_id);
    disable(filename_id);
    disable(tags_id);
    $(buttons_id).hide();
    $(self.id + ' .Deleted').show();
  };

  // 有些事件要在该组件被实体化之后添加才有效。
  self.init = () => {
    const tagsInput = $(self.id + ' .TagsInput');
    const nameInput = $(self.id + ' .NameInput');
    const filename = $(filename_id);

    if (isImage(file.Type)) {
      $(no_link_img_id).hide();
      $(link_img_id).show();
    }

    if (file.Damaged)   $(self.id + ' .Damaged').show();
    if (file.Count > 1) $(file_count_id).show();

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

    $(del_recycle_id).click(() => {
      ajax({method:'POST',url:'/api/delete-file',alerts:ItemAlerts,buttonID:del_btn_id,body:body},
          () => {
            console.log(`文件 [id:${file.ID}] 已扔进回收站，可去回收站找回文件。`);
            self.afterDeleted();
          });
    });

    $(del_btn_id).click(() => {
      disable(del_btn_id);
      ItemAlerts.insert('info', '当删除按钮变红时，再点击一次删除按钮彻底删除该文件，不可恢复。');
      window.setTimeout(() => {
        enable(del_btn_id);
        $(del_btn_id).addClass('text-danger').off().click(() => {
          ajax({method:'POST',url:'/api/really-delete-file',alerts:ItemAlerts,buttonID:del_btn_id,body:body},
          () => {
            self.afterDeleted();
          });
        });
      }, 1000);
    });

    if (file.Deleted) {
      $(tags_btn_id).hide();
      $(restore_btn_id).show();
    } else {
      filename.css('cursor','crosshair').attr({title:'double click to edit'}).dblclick(() => {
        self.toggleFilename();
        nameInput.val(filename.text()).focus();
      });  
    }

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

function delete_file(id) {
  const del_recycle_id = '#f' + id + ' .DeleteRecycle';
  $(del_recycle_id).click();
}
