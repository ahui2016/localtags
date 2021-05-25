
const ImageList = cc('div');

ImageList.append = (files) => {
  files.forEach(file => {
    const item = CreateImageItem(file);
    $(ImageList.id).append(m(item));
    item.init();
  });  
};

ImageList.clear = () => {
  $(ImageList.id).html('');
};

function CreateImageItem(file) {
    if (file.Thumb) {
      file.Thumb = getThumbURL(file.ID);
    } else {
      file.Thumb = getThumbByFiletype(file.Type);
    }
  
    const tagGroup = addPrefix(file.Tags);
    const groupLink = '/light/search?tags=' + encodeURIComponent(tagGroup);
  
    let self = cc('div', file.ID);
  
    self.view = () => m('div').attr({id:self.raw_id}).addClass('card m-1')
      .css('width','8rem').append([
        m('a').addClass('LinkImage').attr({href:getPreviewURL(file.ID),target:'_blank'}).append(
            m('img').attr({title:file.Name,src: file.Thumb}).addClass('card-img-top'),
          ),
        m('div').addClass('ImageButtons card-body p-2 small d-flex justify-content-between').append([
          m('a').text(file.ID).addClass('text-decoration-none')
            .attr({title:'more...',href:'/light/search?fileid='+file.ID, target:'_blank'}),
          m('a').attr({title:addPrefix(file.Tags, '#'), href:groupLink}).append(
            m('i').addClass('bi bi-tag ml-1'),
          ),
          m('i').addClass('DeleteRecycle bi bi-trash ml-1').attr({title:'delete'}),
          m('a').attr({title:'download',download:file.Name, href:getPhotoURL(file.ID)}).append(
            m('i').addClass('bi bi-download ml-1')
          ),
          m('span').text('DELETED').addClass('DELETED badge bg-secondary text-white').hide(),
        ]),
      ]);
  
      self.afterDeleted = () => {
        $(self.id + ' .card-img-top').css('filter', 'opacity(0.5) grayscale(1)');
        disable(self.id + ' .LinkImage');
        $(self.id+' .ImageButtons a').hide();
        $(self.id+' .ImageButtons i').hide();
        $(self.id+' .DELETED').show();
        $(self.id+' .ImageButtons').removeClass('d-flex justify-content-between')
          .addClass('text-center');
      };
  
      self.init = () => {
        const del_btn_id = self.id+' .DeleteRecycle';
  
        const body = new FormData();
        body.append('id', file.ID);
        
        $(del_btn_id).click(() => {
          ajax({method:'POST',url:'/api/delete-file',alerts:Alerts,buttonID:del_btn_id,body:body},
              () => {
                self.afterDeleted();
                Alerts.insert('success', `图片 [id:${self.raw_id}] 已进入回收站，可前往回收站找回。`);
              });
        });
      };
  
    return self;
  }
  