<template>
  <div class="comment-section" id="comments">
    <SubTitle :title="'titles.comment'" />
    <CommentForm />
    <CommentList />
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { SubTitle } from '@/components/Title'
import CommentForm from './CommentForm.vue'
import CommentList from './CommentList.vue'
export default defineComponent({
  name: 'Comment',
  components: { SubTitle, CommentForm, CommentList },
  setup() {
    return {}
  }
})
</script>
<style lang="scss" scoped>
.comment-section {
  background: var(--background-secondary, var(--ob-deep-800));
  padding: 1rem;
  margin-top: 2rem;
  border-radius: 1rem;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.12);
  margin-bottom: 2rem;
}

@media (min-width: 1024px) {
  .comment-section {
    padding: 2.5rem 3.5rem;
    margin-bottom: 0;
  }
}
</style>
